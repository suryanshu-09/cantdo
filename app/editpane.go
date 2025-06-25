package app

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type EditPane struct {
	EditForm *huh.Form
	todo     *Todo
	before   Todo
	edit     *bool
	ready    bool
}

func InitEditPane() {
	edit := new(bool)
	ready := new(bool)
	if app.editpane.todo == nil {
		app.editpane.todo = &Todo{}
	}
	app.editpane = EditPane{
		ready:  *ready,
		todo:   app.editpane.todo,
		before: *app.editpane.todo,
		edit:   edit,
		EditForm: huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					CharLimit(32).
					Title("Change Title").
					Placeholder(app.editpane.before.Title_).
					Value(&app.editpane.todo.Title_).
					Validate(func(t string) error {
						if t == "" {
							return errors.New("sorry but the title cannot be blank")
						}
						return nil
					}).
					WithWidth(38),

				huh.NewText().
					CharLimit(50).
					Title("Change Description").
					Value(&app.editpane.todo.Description_).
					WithWidth(50),

				huh.NewConfirm().
					Title("Edit Todo?").
					Value(edit).
					Affirmative("Yes!").
					Negative("No."),
			).Title("Edit Pane").Description("\n"),
		).WithShowHelp(true),
	}
	// app.editpane.EditForm.WithHeight(termHeight - 6)
}

func (eP EditPane) Init() tea.Cmd {
	return eP.EditForm.Init()
}

func (eP EditPane) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		termHeight = msg.Height
		termWidth = msg.Width
		return eP, eP.Init()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			app.state = viewPane
			return eP, nil
		}
	}
	var cmds []tea.Cmd
	updated, cmd := eP.EditForm.Update(msg)
	if f, ok := updated.(*huh.Form); ok {
		eP.EditForm = f
		cmds = append(cmds, cmd)
	}
	if eP.EditForm.State == huh.StateCompleted {
		if eP.edit != nil && *eP.edit {
			if eP.before.Title_ != eP.todo.Title_ {
				app.db.Delete(eP.before)
				app.db.CreateTodo(*eP.todo)
				app.db.Close()
				app.db, _ = NewDB()
			} else {
				app.db.UpdateDescription(*eP.todo)
			}
			app.Todos = app.db.ReadTodo()
			app.viewpane.InitTodoList(termWidth, termHeight)
		}
		app.state = viewPane
		return app.editpane, nil
	}
	return eP, tea.Batch(cmds...)
}

func (eP EditPane) View() string {
	if termHeight < 40 || termWidth < 80 {
		current := currentStyle.Render(fmt.Sprintf("Current terminal width: %d\nCurrent terminal height: %d\n", termWidth, termHeight))
		shouldBe := shouldBeStyle.Render("Need width: 80\nNeed height: 40")
		warning := lipgloss.JoinVertical(lipgloss.Center, current, shouldBe)
		warningStyle := lipgloss.NewStyle().AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Center).Height(termHeight - 2).Width(termWidth - 2)
		return warningStyle.Render(warning)
	}
	return paneStyle.Height(termHeight - 6).MaxHeight(termHeight - 6).Render(strings.TrimSuffix(eP.EditForm.View(), "\n\n"))
}
