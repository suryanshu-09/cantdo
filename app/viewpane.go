package app

import (
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ViewPane struct {
	Todos list.Model
}

func InitViewPane() {
	vp := ViewPane{}
	vp.Todos = list.Model{}
	vp.InitTodoList(0, 0)
	app.viewpane = vp
}

func (vP ViewPane) Init() tea.Cmd { return nil }

func (vP ViewPane) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		termHeight = msg.Height
		termWidth = msg.Width
		vP.InitTodoList(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			app.state = createPane
			return vP, nil
		}
	}
	var cmd tea.Cmd
	vP.Todos, cmd = vP.Todos.Update(msg)
	return vP, cmd
}

func updateList(msg tea.Msg, m *list.Model) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "delete", "d":
			if !m.IsFiltered() {
				selectedIndex := m.Index()
				if selectedIndex < len(m.Items()) {
					todo := m.SelectedItem().(Todo)
					app.db.Delete(todo)
					app.Todos = app.db.ReadTodo()
					app.viewpane.InitTodoList(termWidth, termHeight)
				}
				items := m.Items()
				items = slices.Delete(items, selectedIndex, selectedIndex+1)
				return m.SetItems(items)
			}
		case "enter", "space":
			if !m.IsFiltered() {
				if len(m.Items()) > 0 {
					selectedIndex := m.Index()
					if selectedIndex < len(m.Items()) {
						todo := m.SelectedItem().(Todo)
						app.db.UpdateStatus(todo)
						if todo.Status_ == PENDING {
							todo.UpdateStatus(COMPLETE)
						} else {
							todo.UpdateStatus(PENDING)
						}
						items := m.Items()
						items[selectedIndex] = todo
						return m.SetItems(items)
					}
				}
			}
		}
	}
	return nil
}

func (vP *ViewPane) InitTodoList(width, height int) {
	del := list.NewDefaultDelegate()
	del.ShowDescription = true
	del.UpdateFunc = updateList
	tl := list.New([]list.Item{}, del, width, height-6)
	tl.Title = "Todos"
	items := make([]list.Item, len(app.Todos))
	for i, t := range app.Todos {
		items[i] = t
	}
	tl.SetItems(items)
	vP.Todos = tl
}

func (vP ViewPane) View() string {
	if termHeight < 24 || termWidth < 80 {
		current := currentStyle.Render(fmt.Sprintf("Current terminal width: %d\nCurrent terminal height: %d\n", termWidth, termHeight))
		shouldBe := shouldBeStyle.Render("Need width: 80\nNeed height: 24")
		warning := lipgloss.JoinVertical(lipgloss.Center, current, shouldBe)
		warningStyle := lipgloss.NewStyle().AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Center).Height(termHeight - 2).Width(termWidth - 2)
		return warningStyle.Render(warning)
	}
	return paneStyle.Render(vP.Todos.View())
}

func (tM *Todo) UpdateStatus(status Status) {
	tM.Status_ = status
}

func (tM Todo) FilterValue() string {
	return tM.Title_
}

func (tM Todo) Title() string {
	return fmt.Sprintf("%s -> %s", tM.Title_, tM.Status())
}

func (tM Todo) Description() string {
	return tM.Description_
}

func (tM Todo) Status() string {
	switch tM.Status_ {
	case PENDING:
		return "Pending"
	case COMPLETE:
		return "Complete"
	case INPROGRESS:
		return "In Progress"
	default:
		return "Pending"
	}
}
