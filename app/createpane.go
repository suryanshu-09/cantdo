package app

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// type CreatePane struct {
// 	titleInput       textinput.Model
// 	descriptionInput textinput.Model
// }
//
// func InitCreatePane() {
// 	cp := CreatePane{}
// 	ti := textinput.New()
// 	cp.titleInput = ti
// 	cp.titleInput.CharLimit = 20
// 	cp.titleInput.VirtualCursor = true
// 	cp.titleInput.Styles.Focused.Prompt = styleFocused
// 	cp.titleInput.Styles.Blurred.Prompt = styleUnfocused
// 	cp.titleInput.Styles.Cursor.Shape = tea.CursorBlock
// 	cp.titleInput.Styles.Cursor.Blink = true
// 	cp.titleInput.Styles.Cursor.BlinkSpeed = time.Second * 1
// 	cp.titleInput.SetWidth(20)
// 	cp.titleInput.Prompt = "Create Title: "
// 	cp.titleInput.Placeholder = "Title"
// 	cp.titleInput.Focus()
//
// 	di := textinput.New()
// 	cp.descriptionInput = di
// 	cp.descriptionInput.CharLimit = 50
// 	cp.descriptionInput.VirtualCursor = true
// 	cp.descriptionInput.Styles.Focused.Prompt = styleFocused
// 	cp.descriptionInput.Styles.Blurred.Prompt = styleUnfocused
// 	cp.descriptionInput.Styles.Cursor.Shape = tea.CursorBlock
// 	cp.descriptionInput.Styles.Cursor.Blink = true
// 	cp.descriptionInput.Styles.Cursor.BlinkSpeed = time.Second * 1
// 	cp.descriptionInput.SetWidth(50)
// 	cp.descriptionInput.Prompt = "Create Description: "
// 	cp.descriptionInput.Placeholder = "Description"
// 	createpane = cp
// }
//
// func (cP CreatePane) Init() tea.Cmd { return nil }
//
// func (cP CreatePane) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	var cmd tea.Cmd
// 	var cmds []tea.Cmd
//
// 	switch msg := msg.(type) {
// 	case tea.WindowSizeMsg:
// 		termHeight = msg.Height
// 		termWidth = msg.Width
// 	case tea.KeyMsg:
// 		switch msg.String() {
// 		case "ctrl+c":
// 			return cP, tea.Quit
// 		case "tab":
// 			state = editPane
// 			return cP, nil
// 		case "up", "down":
// 			if cP.titleInput.Focused() {
// 				cP.titleInput.Blur()
// 				cmd = cP.descriptionInput.Focus()
// 				cmds = append(cmds, cmd)
// 			} else {
// 				cP.descriptionInput.Blur()
// 				cmd = cP.titleInput.Focus()
// 				cmds = append(cmds, cmd)
// 			}
// 		case "enter":
// 			title := cP.titleInput.Value()
// 			desc := cP.descriptionInput.Value()
// 			if title != "" && desc != "" {
// 				todo := Todo{Title_: title, Description_: desc}
// 				db.CreateTodo(todo)
// 				Todos = db.ReadTodo()
// 				viewpane.InitTodoList(termWidth, termHeight)
// 				cP.titleInput.Reset()
// 				cP.descriptionInput.Reset()
// 				cmd = cP.titleInput.Focus()
// 				cmds = append(cmds, cmd)
// 				cP.descriptionInput.Blur()
// 			}
// 		}
// 	}
//
// 	cP.titleInput, cmd = cP.titleInput.Update(msg)
// 	cmds = append(cmds, cmd)
//
// 	cP.descriptionInput, cmd = cP.descriptionInput.Update(msg)
// 	cmds = append(cmds, cmd)
// 	return cP, tea.Batch(cmds...)
// }
//
// func (cP CreatePane) View() string {
// 	if termHeight < 24 || termWidth < 80 {
// 		current := currentStyle.Render(fmt.Sprintf("Current terminal width: %d\nCurrent terminal height: %d\n", termWidth, termHeight))
// 		shouldBe := shouldBeStyle.Render("Need width: 80\nNeed height: 24")
// 		warning := lipgloss.JoinVertical(lipgloss.Center, current, shouldBe)
// 		warningStyle := lipgloss.NewStyle().AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Center).Height(termHeight - 2).Width(termWidth - 2)
// 		return warningStyle.Render(warning)
// 	}
// 	return paneStyle.Render(lipgloss.JoinVertical(lipgloss.Left, cP.titleInput.View(), cP.descriptionInput.View()))
// }

type CreatePane struct {
	CreateForm *huh.Form
}

func InitCreatePane() {
	todo := Todo{}
	var create bool
	app.createpane = CreatePane{
		CreateForm: huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					CharLimit(32).
					Title("New Title For Your Todo").
					Placeholder("Title").
					Description("Enter a title").
					Value(&todo.Title_).
					Validate(func(t string) error {
						if t == "" {
							return errors.New("Sorry but the title cannot be blank")
						}
						return nil
					}).
					WithWidth(32),

				huh.NewText().
					CharLimit(50).
					Title("New Description For Your Todo").
					Placeholder("Description").
					Description("Enter a description").
					Value(&todo.Description_).
					WithWidth(50),

				huh.NewConfirm().
					Title("Create Todo?").
					Value(&create).
					Affirmative("Yes!").
					Negative("No."),
			),
		).WithShowHelp(true),
	}
	if create {
		app.db.CreateTodo(todo)
	}
}

func (cP CreatePane) Init() tea.Cmd { return cP.CreateForm.Init() }

func (cP CreatePane) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		termHeight = msg.Height
		termHeight = msg.Width
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
		cmds = append(cmds, tea.Quit)
	}
	return cP, tea.Batch(cmds...)
}

func (cP CreatePane) View() string {
	// return paneStyle.Render(strings.TrimSuffix(cP.CreateForm.View(), "\n\n"))
	return cP.CreateForm.View()
}

func RunHuh() {
	InitCreatePane()
	_, err := tea.NewProgram(app.createpane, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
}
