package docs

import (
	"fmt"
	"io/fs"
	"strings"

	"codeberg.org/splitringresonator/multiband/docs"
	"codeberg.org/splitringresonator/multiband/internal/version"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Margin(1, 2).
			AlignHorizontal(lipgloss.Center).
			Background(lipgloss.Color("5"))
	headerStyle = lipgloss.NewStyle().
			Margin(1, 2).
			AlignHorizontal(lipgloss.Center).
			Foreground(lipgloss.Color("5"))
	errorTextStyle = lipgloss.NewStyle().
			Margin(2, 2).
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Foreground(lipgloss.Color("9"))
		//BorderStyle(lipgloss.RoundedBorder()).
		//BorderForeground(lipgloss.Color("62"))
	//
	//PaddingRight(2)
	//	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	//	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle       = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

type Origin uint8

const (
	OriginDefault Origin = iota
	OriginEmbedded
	OriginRemote
)

type item struct {
	title                      string
	origin                     Origin
	content, snippet, revision *string

	err error
}

func (i item) FilterValue() string {
	if i.content != nil {
		return i.title + *i.content
	}

	if i.snippet != nil {
		return i.title + *i.snippet
	}

	return i.title
}

func (i item) Description() string {
	if i.err != nil {
		return i.err.Error()
	}
	if i.snippet != nil {
		return *i.snippet
	}
	return ""
}

func (i item) Title() string {
	trinkets := []string{}

	/*
		switch origin := i.origin; origin {
		case OriginEmbedded:
			trinkets = append(trinkets, "λ")

		case OriginRemote:
			trinkets = append(trinkets, "remote")
		}
	*/

	if len(trinkets) > 0 {
		return fmt.Sprintf("%s (%s)", i.title, strings.Join(trinkets, " "))
	} else {
		return i.title
	}
}

/* if we want to control how each item is represented in the list, use this starter
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
*/

type Model struct {
	choice     string
	history    []string
	rawContent string
	quitting   bool
	width      uint
	height     uint

	list     list.Model
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

	renderer, err := m.getRenderer()

	if err != nil {
		return errorTextStyle.Render(err.Error())
	}

	str, err := renderer.Render(m.rawContent)
	if err != nil {
		return errorTextStyle.Render(err.Error())
	}

	return str
}

func (m *Model) updateWindowSize(msg tea.WindowSizeMsg) {
	m.viewport.Style.Height(msg.Height)
	m.list.SetHeight(msg.Height)
	m.list.SetWidth(msg.Width)
	m.width = uint(msg.Width)
	m.height = uint(msg.Height)

	if m.choice != "" {
		// adjust width of viewport or rendered content
		m.viewport.SetContent(m.renderContent())
	}
}

func (m *Model) loadSelection(path string) error {
	m.choice = path
	m.history = append(m.history, m.choice)

	m.rawContent = string(*m.list.SelectedItem().(item).content)
	m.viewport.SetContent(m.renderContent())

	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.updateWindowSize(msg)

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			// Don't match any of the keys below if we're actively filtering in the list
			break
		}

		//if m.choice != "" {
		// if the pager is visible, do not propagate events to other components
		//break
		//}

		switch keypress := msg.String(); keypress {

		case "esc":
			if m.choice != "" {
				m.choice = ""
				// do not propagate event down to pager, to avoid exiting
				return m, nil
			}

		case "ctrl+f":
			m.viewport.HalfPageDown()

		case "ctrl+b":
			m.viewport.HalfPageUp()

		case "q", "ctrl+c":
			m.quitting = true
			cmds = append(cmds, tea.Quit)

		case "enter":
			if m.choice == "" {
				if i, ok := m.list.SelectedItem().(item); ok {
					m.loadSelection(i.title)
				}
			}

		}
	}

	if m.choice != "" {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.choice != "" {
		pct := fmt.Sprintf("%.0f%%", (float32(m.viewport.YOffset+m.viewport.VisibleLineCount())/float32(m.viewport.TotalLineCount()))*100.0)
		ln := fmt.Sprintf("ln:%d:%d:%d", m.viewport.YOffset, m.viewport.YOffset+m.viewport.VisibleLineCount(), m.viewport.TotalLineCount())
		return headerStyle.Render(fmt.Sprintf("> %s@%s (%s %s)", m.choice, version.Build, pct, ln)) + "\n" + m.viewport.View()
	}

	if m.quitting {
		return strings.Join(m.history, "\n")
	}

	return "\n" + m.list.View()
}

func NewModel(width uint, height uint) Model {
	if width == 0 {
		width = 78
	}
	if height == 0 {
		height = 20
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

			dat, err := docs.Docs.ReadFile(path)
			var content, snippet string
			if dat != nil {
				content = string(dat)
				snipLength := 60
				if len(content) < snipLength {
					snippet = content
				} else {
					snippet = strings.ReplaceAll(content, "\n", " ")[:snipLength] + "..."
				}
			}

			items = append(items, item{
				title:    path,
				revision: &version.Build,
				origin:   OriginEmbedded,
				snippet:  &snippet,
				content:  &content,
				err:      err,
			})
		}
		return nil
	})

	l := list.New(items, list.NewDefaultDelegate(), int(width), int(height))
	l.Title = fmt.Sprintf("Multiband Embedded Documentation Browser %s", version.Build)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	l.SetWidth(int(width))
	l.SetHeight(int(height))

	vp := viewport.New(int(width), int(height-(uint(headerStyle.GetHeight()))))
	vp.Style = lipgloss.NewStyle()
	//BorderStyle(lipgloss.RoundedBorder()).
	//BorderForeground(lipgloss.Color("62")).
	//PaddingRight(2)

	return Model{
		width:    width,
		height:   height,
		viewport: vp,
		list:     l,
		history:  []string{},
	}
}
