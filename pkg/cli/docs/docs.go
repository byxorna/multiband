package docs

import (
	"fmt"
	"io"
	"io/fs"
	"strings"

	"codeberg.org/splitringresonator/multiband/docs"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
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

func (i item) FilterValue() string { return string(i) }

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

type Model struct {
	list     list.Model
	choice   string
	quitting bool
	width    uint

	viewport viewport.Model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) getRenderer() (*glamour.TermRenderer, error) {
	const glamourGutter = 2
	glamourRenderWidth := int(m.width) - m.viewport.Style.GetHorizontalFrameSize() - glamourGutter

	return glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithColorProfile(lipgloss.ColorProfile()),
		glamour.WithWordWrap(glamourRenderWidth),
		glamour.WithWordWrap(int(m.width)), //nolint:gosec
		glamour.WithBaseURL(m.choice),
		glamour.WithPreservedNewLines(),
	)
}

func (m Model) renderContent() string {
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
	renderer, err := m.getRenderer()

	if err != nil {
		return quitTextStyle.Render(err.Error())
	}

	str, err := renderer.Render(string(dat))
	if err != nil {
		return quitTextStyle.Render(err.Error())
	}

	return str
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.width = uint(msg.Width)

		if m.choice != "" {
			// adjust width of viewport or rendered content
			m.viewport.SetContent(m.renderContent())
		}

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch keypress := msg.String(); keypress {

		case "esc":
			if m.choice != "" {
				m.choice = ""
			}

		case "ctrl+f":
			m.viewport.HalfPageDown()

		case "ctrl+b":
			m.viewport.HalfPageUp()

		case "q", "ctrl+c":
			m.quitting = true
			cmds = append(cmds, tea.Quit)

		case "enter":
			if i, ok := m.list.SelectedItem().(item); ok {
				m.choice = string(i)
			}

			m.viewport.SetContent(m.renderContent())

		default:
			//if m.list.ShowFilter() {
			//	m.filter += strings.ToLower(msg.String())
			//	m.list.SetFilterText(m.filter)
			//}

		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	if m.choice != "" {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.choice != "" {

		return fmt.Sprintf("\n > %s\n\n", m.choice) + m.viewport.View()
	}

	if m.quitting {
		return quitTextStyle.Render("Ta!")
	}

	return "\n" + m.list.View()
}

func NewModel(width uint) Model {
	if width == 0 {
		width = 78
	}

	items := []list.Item{}

	fs.WalkDir(docs.Docs, ".", func(path string, d fs.DirEntry, err error) error {
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
	})

	// TODO: can we fzf through contents from this view?

	l := list.New(items, itemDelegate{}, int(width), listHeight)
	l.Title = "Embedded Documentation Browser"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	l.SetWidth(int(width))

	vp := viewport.New(int(width), 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	return Model{
		width:    width,
		viewport: vp,
		list:     l,
	}
}
