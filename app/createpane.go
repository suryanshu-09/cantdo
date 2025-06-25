package app

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type CreatePane struct {
	CreateForm *huh.Form
	todo       *Todo
	create     *bool
}

func InitCreatePane() {
	todo := &Todo{}
	create := new(bool)
	app.createpane = CreatePane{
		todo:   todo,
		create: create,
		CreateForm: huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					CharLimit(32).
					Title("New Title For Your Todo").
					Placeholder("Title").
					Value(&todo.Title_).
					Validate(func(t string) error {
						if t == "" {
							return errors.New("sorry but the title cannot be blank")
						}
						return nil
					}).
					WithWidth(38),

				huh.NewText().
					CharLimit(50).
					Title("New Description For Your Todo").
					Placeholder("Description").
					Value(&todo.Description_).
					WithWidth(50),

				huh.NewConfirm().
					Title("Create Todo?").
					Value(create).
					Affirmative("Yes!").
					Negative("No."),
			).Title("Create Pane").Description("\n"),
		).WithShowHelp(true),
	}
}

func (cP CreatePane) Init() tea.Cmd { return cP.CreateForm.Init() }

func (cP CreatePane) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		termHeight = msg.Height
		termWidth = msg.Width
		return cP, cP.Init()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return cP, tea.Quit
		case "tab":
			app.state = viewPane
			return cP, nil
		}
	}
	var cmds []tea.Cmd
	updated, cmd := cP.CreateForm.Update(msg)
	if f, ok := updated.(*huh.Form); ok {
		cP.CreateForm = f
		cmds = append(cmds, cmd)
	}
	if cP.CreateForm.State == huh.StateCompleted {
		if cP.create != nil && *cP.create {
			app.db.CreateTodo(*cP.todo)
			app.Todos = app.db.ReadTodo()
			app.viewpane.InitTodoList(termWidth, termHeight)
		}
		InitCreatePane()
		app.state = viewPane
		return app.createpane, app.createpane.Init()
	}
	return cP, tea.Batch(cmds...)
}

func (cP CreatePane) View() string {
	if termHeight < 40 || termWidth < 80 {
		current := currentStyle.Render(fmt.Sprintf("Current terminal width: %d\nCurrent terminal height: %d\n", termWidth, termHeight))
		shouldBe := shouldBeStyle.Render("Need width: 80\nNeed height: 40")
		warning := lipgloss.JoinVertical(lipgloss.Center, current, shouldBe)
		warningStyle := lipgloss.NewStyle().AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Center).Height(termHeight - 2).Width(termWidth - 2)
		return warningStyle.Render(warning)
	}
	return paneStyle.Height(termHeight - 6).MaxHeight(termHeight - 6).Render(strings.TrimSuffix(cP.CreateForm.View(), "\n\n"))
}
