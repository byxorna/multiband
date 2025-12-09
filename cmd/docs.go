package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"codeberg.org/splitringresonator/multiband/docs"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s", i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type docsModel struct {
	list     list.Model
	choice   string
	quitting bool
	width    uint

	viewport viewport.Model
}

func (m docsModel) Init() tea.Cmd {
	return nil
}

func (m docsModel) renderContent() string {
	if m.choice == "" {
		return ""
	}

	dat, err := docs.Docs.ReadFile(m.choice)
	if err != nil {
		return quitTextStyle.Render(err.Error())
	}

	//	quitTextStyle.Render(fmt.Sprintf("❱ %s (%d bytes)\n\n", m.choice, len(dat)) + string(dat) + "\n")

	// We need to adjust the width of the glamour render from our main width
	// to account for a few things:
	//
	//  * The viewport border width
	//  * The viewport padding
	//  * The viewport margins
	//  * The gutter glamour applies to the left side of the content
	//
	const glamourGutter = 2
	glamourRenderWidth := int(m.width) - m.viewport.Style.GetHorizontalFrameSize() - glamourGutter

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithColorProfile(lipgloss.ColorProfile()),
		glamour.WithWordWrap(glamourRenderWidth),
		glamour.WithWordWrap(int(m.width)), //nolint:gosec
		glamour.WithBaseURL(m.choice),
		glamour.WithPreservedNewLines(),
	)
	if err != nil {
		return quitTextStyle.Render(err.Error())
	}

	str, err := renderer.Render(string(dat))
	if err != nil {
		return quitTextStyle.Render(err.Error())
	}

	return str
}

func (m docsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.width = uint(msg.Width)

		if m.choice != "" {
			// adjust width of viewport or rendered content
			m.viewport.SetContent(m.renderContent())
		}
		m.viewport, cmd = m.viewport.Update(msg)

		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}

			m.viewport.SetContent(m.renderContent())

			return m, nil

		}
	default:
		return m, nil
	}

	if m.choice == "" {
		m.list, cmd = m.list.Update(msg)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

func (m docsModel) View() string {
	if m.choice != "" {
		return m.viewport.View()
	}

	if m.quitting {
		return quitTextStyle.Render("Ta!")
	}

	return "\n" + m.list.View()
}

func newDocsModel(width uint, l list.Model) docsModel {
	if width == 0 {
		width = 78
	}

	l.SetWidth(int(width))

	vp := viewport.New(int(width), 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	return docsModel{
		width:    width,
		viewport: vp,
		list:     l,
	}
}

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "View built-in documentation",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		items := []list.Item{}

		if err := fs.WalkDir(docs.Docs, ".", func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				name := d.Name()

				if name == "." || name == ".." {
					return nil
				}

				return nil // keep walking
			}

			if !d.IsDir() {
				if !strings.HasSuffix(d.Name(), ".md") {
					return nil
				}

				items = append(items, item(path))
			}
			return nil
		}); err != nil {
			return err
		}

		var width uint
		isTerminal := term.IsTerminal(int(os.Stdout.Fd()))
		if isTerminal {
			w, _, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				width = uint(w) //nolint:gosec
			}

			if width > 120 {
				width = 120
			}
		}
		if width == 0 {
			width = 80
		}

		// TODO: can we fzf through contents from this view?

		l := list.New(items, itemDelegate{}, int(width), listHeight)
		l.Title = "Documentation browser"
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		l.Styles.Title = titleStyle
		l.Styles.PaginationStyle = paginationStyle
		l.Styles.HelpStyle = helpStyle

		p := tea.NewProgram(newDocsModel(width, l))

		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}
