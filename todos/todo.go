package todos

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/v2/textinput"
	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type ListTodo struct {
	Title       string
	Description string
	Status      Status
}

type (
	CreatePane struct {
		cursor int
		inputs []textinput.Model
	}
	EditPane struct {
		cursorTodo int
		cursorInp  int
		inputs     [][]textinput.Model
	}
	ViewPane struct {
		cursor   int
		selected map[int]bool
	}
)

type PaneModel struct {
	viewPort viewport.Model
	body     string
	db       *DB
	Focus    int
	ToDos    []ListTodo
	EP       *EditPane
	VP       *ViewPane
	CP       *CreatePane
}

var (
	styleFocused      = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff69b4")).Bold(true)
	styleUnfocused    = lipgloss.NewStyle().Foreground(lipgloss.Color("#606060"))
	styleSelected     = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	styleTitle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6347"))
	styleDesc         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#2e8b57"))
	styleStatus       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#DAA520"))
	styleHint         = lipgloss.NewStyle().Foreground(lipgloss.Color("#606060"))
	styleHighlightTab = lipgloss.NewStyle().Bold(true).Italic(true).Foreground(lipgloss.Color("#202020")).Background(lipgloss.Color("#69ffb4"))
	styleCursor       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8a2be2"))
	focusedButton     = styleFocused.Render("[ Submit ]")
	blurredButton     = fmt.Sprintf("[ %s ]", styleUnfocused.Render("Submit"))
	epButton          = false
	terminalWidth     int
	terminalHeight    int
	ready             = false
)

func CreateEditPane(todos []ListTodo) [][]textinput.Model {
	editTodos := make([][]textinput.Model, 0)
	for _, todo := range todos {
		editTitle := textinput.New()
		editTitle.Placeholder = todo.Title
		editTitle.Prompt = "Edit Title: "
		editTitle.SetWidth(128)
		editTitle.CharLimit = 32

		editTitle.Styles.Focused.Prompt = styleFocused
		editTitle.Styles.Blurred.Text = styleUnfocused

		editDesc := textinput.New()
		editDesc.Placeholder = todo.Description
		editDesc.Prompt = "Edit Description: "
		editDesc.SetWidth(160)
		editDesc.CharLimit = 128

		editDesc.Styles.Focused.Prompt = styleFocused
		editDesc.Styles.Blurred.Text = styleUnfocused

		editTodos = append(editTodos, []textinput.Model{editTitle, editDesc})
	}
	if len(todos) > 0 {
		editTodos[0][0].Focus()
	}
	return editTodos
}

func newSession(todosList []ListTodo) *PaneModel {
	todos := make([]ListTodo, 0)
	titleInput := textinput.New()
	titleInput.Placeholder = "Title"
	titleInput.Prompt = "Title: "
	titleInput.SetWidth(44)
	titleInput.CharLimit = 32

	titleInput.Styles.Focused.Prompt = styleFocused
	titleInput.Styles.Blurred.Text = styleUnfocused
	titleInput.Focus()

	descInput := textinput.New()
	descInput.Placeholder = "Description"
	descInput.Prompt = "Description: "
	descInput.SetWidth(160)
	descInput.CharLimit = 128

	descInput.Styles.Focused.Prompt = styleFocused
	descInput.Styles.Blurred.Text = styleUnfocused

	db, _ := NewDB()
	for _, todo := range todosList {
		db.CreateTodo(todo)
	}
	todosDB := db.ReadTodo()
	todos = append(todos, todosDB...)
	return &PaneModel{
		db:    db,
		ToDos: todos,
		VP: &ViewPane{
			cursor:   0,
			selected: make(map[int]bool),
		},
		EP: &EditPane{
			cursorTodo: 0,
			cursorInp:  0,
			inputs:     CreateEditPane(todos),
		},
		CP: &CreatePane{
			cursor: 0,
			inputs: []textinput.Model{titleInput, descInput},
		},
	}
}

func (pM PaneModel) Init() tea.Cmd {
	return textinput.Blink
}

func (pM PaneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerViewHeight := lipgloss.Height(pM.headerView())
		footerViewHeight := lipgloss.Height(pM.footerView())
		terminalWidth = msg.Width
		terminalHeight = msg.Height
		if !ready {
			vp := viewport.New(viewport.WithHeight(msg.Height-headerViewHeight-footerViewHeight), viewport.WithWidth(msg.Width))
			vp.SoftWrap = false
			vp.YPosition = headerViewHeight
			pM.viewPort = vp
			ready = true
		} else {
			vp := viewport.New(viewport.WithHeight(msg.Height-headerViewHeight-footerViewHeight), viewport.WithWidth(msg.Width))
			pM.viewPort = vp
		}
		pM.viewPort.SetContent(pM.body)
		pM.updateViewportContent()
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+x":
			pM.db.Flush()
			return pM, nil

		case "ctrl+c":
			return pM, tea.Quit

		case "q":
			if pM.Focus == 0 {
				pM.db.Close()
				return pM, tea.Quit
			}

		case "left":
			pM.Focus = (pM.Focus - 1 + 3) % 3
			pM.updateViewportContent()
			return pM, nil

		case "right":
			pM.Focus = (pM.Focus + 1) % 3
			pM.updateViewportContent()
			return pM, nil

		case "down":
			if len(pM.ToDos) > 0 {
				switch pM.Focus {
				case 2:
					pM.CP.inputs[pM.CP.cursor].Blur()
					pM.CP.cursor = (pM.CP.cursor + 1) % 2
					pM.CP.inputs[pM.CP.cursor].Focus()
				case 1:
					if pM.EP.cursorTodo < len(pM.EP.inputs)-1 || (pM.EP.cursorTodo == len(pM.EP.inputs)-1 && pM.EP.cursorInp == 0) {
						pM.EP.inputs[pM.EP.cursorTodo][pM.EP.cursorInp].Blur()
						if pM.EP.cursorInp == 1 {
							pM.EP.cursorTodo = pM.EP.cursorTodo + 1
						}
						pM.EP.cursorInp = (pM.EP.cursorInp + 1) % 2
						pM.EP.inputs[pM.EP.cursorTodo][pM.EP.cursorInp].Focus()
					}
				case 0:
					// pM.VP.cursor = (pM.VP.cursor + 1) % len(pM.ToDos)
					if pM.VP.cursor < len(pM.ToDos)-1 {
						pM.VP.cursor++
					}
				}
			}
			pM.updateViewportContent()

		case "up":
			if len(pM.ToDos) > 0 {
				switch pM.Focus {
				case 2:
					pM.CP.inputs[pM.CP.cursor].Blur()
					pM.CP.cursor = (pM.CP.cursor + 1) % 2
					pM.CP.inputs[pM.CP.cursor].Focus()
				case 1:
					if pM.EP.cursorTodo > 0 || (pM.EP.cursorTodo == 0 && pM.EP.cursorInp == 1) {
						pM.EP.inputs[pM.EP.cursorTodo][pM.EP.cursorInp].Blur()
						if pM.EP.cursorInp == 0 {
							pM.EP.cursorTodo = pM.EP.cursorTodo - 1
						}
						pM.EP.cursorInp = (pM.EP.cursorInp + 1) % 2
						pM.EP.inputs[pM.EP.cursorTodo][pM.EP.cursorInp].Focus()
					}
				case 0:
					// pM.VP.cursor = (pM.VP.cursor - 1 + len(pM.ToDos)) % len(pM.ToDos)
					if pM.VP.cursor > 0 {
						pM.VP.cursor--
					}
				}
			}
			pM.updateViewportContent()

		case "space":
			// if pM.Focus == 0 {
			// 	pM.VP.selected[pM.VP.cursor] = !pM.VP.selected[pM.VP.cursor]
			// }
			// pM.updateViewportContent()

		case "enter":
			switch pM.Focus {
			case 2:
				if pM.CP.cursor == len(pM.CP.inputs)-1 {
					title := strings.TrimSpace(pM.CP.inputs[0].Value())
					desc := strings.TrimSpace(pM.CP.inputs[1].Value())
					if title != "" && desc != "" {
						pM.db.CreateTodo(ListTodo{Title: title, Description: desc, Status: PENDING})
						pM.ToDos = pM.db.ReadTodo()
						for i := range pM.CP.inputs {
							pM.CP.inputs[i].SetValue("")
						}
						pM.CP.cursor = 0
						pM.CP.inputs[0].Focus()
						pM.CP.inputs[1].Blur()
					}
					epButton = false
					pM.EP.inputs = CreateEditPane(pM.ToDos)
					pM.updateViewportContent()
					return pM, nil
				}

			case 1:
				epButton = false
				for i, inputs := range pM.EP.inputs {
					if inputs[0].Value() != "" {
						pM.ToDos[i].Title = inputs[0].Value()
					}
					if inputs[1].Value() != "" {
						pM.ToDos[i].Description = inputs[1].Value()
					}

					pM.EP.inputs[pM.EP.cursorTodo][pM.EP.cursorInp].Blur()
					pM.EP.cursorInp = 0
					pM.EP.cursorTodo = 0
					pM.EP.inputs[0][0].Focus()
				}
				pM.db.UpdateTodo(pM.ToDos)
				pM.updateViewportContent()

			case 0:
				for i, selected := range pM.ToDos {
					if pM.VP.selected[i] {
						pM.db.UpdateStatus(selected)
					}
				}
				pM.ToDos = pM.db.ReadTodo()
				pM.updateViewportContent()
				return pM, nil
			}

		case "tab":
			switch pM.Focus {
			case 2:
				pM.CP.inputs[pM.CP.cursor].Blur()
				pM.CP.cursor = (pM.CP.cursor + 1) % 2
				pM.CP.inputs[pM.CP.cursor].Focus()

			case 1:
				// pM.EP.inputs[pM.EP.cursorTodo][pM.EP.cursorInp].Blur()
				// if pM.EP.cursorInp == 1 {
				// 	pM.EP.cursorTodo = (pM.EP.cursorTodo + 1) % len(pM.EP.inputs)
				// }
				// pM.EP.cursorInp = (pM.EP.cursorInp + 1) % 2
				// pM.EP.inputs[pM.EP.cursorTodo][pM.EP.cursorInp].Focus()

			case 0:
				pM.VP.selected[pM.VP.cursor] = !pM.VP.selected[pM.VP.cursor]
				pM.VP.selected = make(map[int]bool, 0)
			}
			pM.updateViewportContent()
		}
	}

	for i := range pM.CP.inputs {
		var cmd tea.Cmd
		pM.CP.inputs[i], cmd = pM.CP.inputs[i].Update(msg)
		pM.updateViewportContent()
		cmds = append(cmds, cmd)
	}

	for i := range pM.EP.inputs {
		var cmd tea.Cmd
		for j := range pM.EP.inputs[i] {
			pM.EP.inputs[i][j], cmd = pM.EP.inputs[i][j].Update(msg)
			pM.updateViewportContent()
			epButton = true
			cmds = append(cmds, cmd)
		}
	}

	var cmd tea.Cmd
	pM.viewPort, cmd = pM.viewPort.Update(msg)
	cmds = append(cmds, cmd)

	return pM, tea.Batch(cmds...)
}

func (pM *PaneModel) updateViewportContent() {
	var b strings.Builder

	switch pM.Focus {
	case 0:
		if len(pM.ToDos) == 0 {
			b.WriteString(styleHint.Render("Go to create pane to create a new todo."))
			b.WriteString("\n")
		}
		for i, todo := range pM.ToDos {
			cursor := "  "
			if pM.VP.cursor == i {
				cursor = styleCursor.Render("> ")
			}
			checked := "[ ]"
			if pM.VP.selected[i] {
				checked = styleSelected.Render("[x]")
			}
			b.WriteString(fmt.Sprintf("%s%s %s %s\n", cursor, checked, styleTitle.Render(todo.Title), styleDesc.Render("→ "+todo.Description)))
			b.WriteString("      " + styleStatus.Render(fmt.Sprintf("Status: %d", todo.Status)))
			if i != len(pM.ToDos)-1 {
				b.WriteString("\n\n")
			}
		}
	case 1:
		if len(pM.EP.inputs) == 0 {
			b.WriteString(styleHint.Render("Edit pane empty, please add some todos in the create pane.\n"))
			b.WriteString("\n")
		}
		for i, inputArr := range pM.EP.inputs {
			ind := fmt.Sprintf("%3d. ", i+1)
			for i, input := range inputArr {
				if i == 0 {
					b.WriteString(ind)
					b.WriteString(input.View() + "\n")
				} else {
					space := strings.Repeat(" ", len(ind))
					b.WriteString(space + input.View() + "\n")
				}
			}
			b.WriteString("\n")
		}
		if epButton {
			b.WriteString(focusedButton)
		} else {
			b.WriteString(blurredButton)
		}
	case 2:
		for _, input := range pM.CP.inputs {
			b.WriteString(input.View() + "\n\n")
		}
		if pM.CP.inputs[1].Value() != "" {
			b.WriteString(focusedButton)
		} else {
			b.WriteString(blurredButton)
		}
	}

	headerViewHeight := lipgloss.Height(pM.headerView())
	footerViewHeight := lipgloss.Height(pM.footerView())
	borderHeight := terminalHeight - headerViewHeight - footerViewHeight
	finalBody := lipgloss.NewStyle().Padding(1, 2).Align(lipgloss.Left).Border(lipgloss.RoundedBorder(), false, true).Width(terminalWidth).Height(borderHeight).Render(b.String())

	pM.body = finalBody
	pM.viewPort.SetContent(pM.body)
}

func (pM PaneModel) FindHeigthtEP(cursorTodo, cursorInp int) int {
	height := 0
	j := 0
	for i := range cursorTodo + 1 {
		if i == cursorTodo {
			height += lipgloss.Height(pM.EP.inputs[i][j].Prompt)
		}
		height += lipgloss.Height(pM.EP.inputs[i][j].Prompt)
		j = (j + 1) % 2
	}
	return height - 1
}

func (pM PaneModel) FindHeigthtCP(c int) int {
	height := 0
	for i := range c + 1 {
		height += lipgloss.Height(pM.CP.inputs[i].Prompt) + i
	}
	return height
}

func (pM PaneModel) View() (string, *tea.Cursor) {
	var cursor *tea.Cursor

	headerViewHeight := lipgloss.Height(pM.headerView())
	footerViewHeight := lipgloss.Height(pM.footerView())
	borderHeight := terminalHeight - headerViewHeight - footerViewHeight
	switch pM.Focus {
	case 1:
		if len(pM.EP.inputs) > 0 {
			cursor = pM.EP.inputs[pM.EP.cursorTodo][pM.EP.cursorInp].Cursor()
			cursor.Y = pM.FindHeigthtEP(pM.EP.cursorTodo, pM.EP.cursorInp) + lipgloss.Height(pM.headerView())%(borderHeight)
			cursor.X = lipgloss.Width(pM.EP.inputs[pM.EP.cursorTodo][pM.EP.cursorInp].Value()+pM.EP.inputs[pM.EP.cursorTodo][pM.EP.cursorInp].Prompt) + 8
			cursor.Blink = true
			cursor.Color = lipgloss.Color("#69ffb4")
		}
	case 2:
		cursor = pM.CP.inputs[pM.CP.cursor].Cursor()
		cursor.Y = pM.FindHeigthtCP(pM.CP.cursor) + lipgloss.Height(pM.headerView())
		cursor.X = lipgloss.Width(pM.CP.inputs[pM.CP.cursor].Value()+pM.CP.inputs[pM.CP.cursor].Prompt) + 3
		cursor.Blink = true
		cursor.Color = lipgloss.Color("#69ffb4")
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		pM.headerView(),
		pM.viewPort.View(),
		pM.footerView()), cursor
}

func (pM PaneModel) headerView() string {
	renderTab := func(label string, active bool) string {
		if active {
			return styleHighlightTab.Render(label)
		}
		return styleHint.Render(label)
	}
	title := fmt.Sprintf("%s    %s    %s",
		renderTab("View Todos", pM.Focus == 0),
		renderTab("Edit Todos", pM.Focus == 1),
		renderTab("Create Todos", pM.Focus == 2),
	)
	headerStyle := lipgloss.RoundedBorder()
	headerStyle.Right = "├"
	hStyle := lipgloss.NewStyle().BorderStyle(headerStyle)
	text := hStyle.Render(title)
	line := strings.Repeat("─", max(0, terminalWidth-lipgloss.Width(text)))
	header := lipgloss.JoinHorizontal(lipgloss.Center, text, line)
	return header
}

func (pM PaneModel) footerView() string {
	quit := (styleHint.Render("ctrl+c: Quit"))
	flush := ("    " + styleHint.Render("ctrl+x: Flush Data"))
	var hints string
	switch pM.Focus {
	case 2:
		hints = ("    " + styleHint.Render("tab: next field | enter: submit"))
	case 1:
		hints = ("    " + styleHint.Render("arrow keys: move around | enter: submit"))
	case 0:
		hints = ("    " + styleHint.Render("tab: select | enter: update status"))
	}
	footerStyle := lipgloss.RoundedBorder()
	footerStyle.Left = "┤"
	fStyle := lipgloss.NewStyle().BorderStyle(footerStyle)
	text := fStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center, quit, flush, hints))
	line := strings.Repeat("─", max(0, terminalWidth-lipgloss.Width(text)))
	footer := lipgloss.JoinHorizontal(lipgloss.Center, line, text)
	return footer
}

func Render() {
	// todos := []ListTodo{{Title: "Gae Man tries Golang", Description: "Let's see if he's any good.", Status: PENDING}, {Title: "Trying harder he is", Description: "Won't make it he.", Status: PENDING}}
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()
	if _, err := tea.NewProgram(newSession(nil), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		slog.Error("could not run app", "error", err)
		os.Exit(1)
	}
}
