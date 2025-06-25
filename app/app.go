package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type State int

const (
	viewPane State = iota
	createPane
	pomoDoro
	editPane
)

var (
	termHeight = 0 // 24
	termWidth  = 0 // 84
	app        = &AppModel{}
)

type AppModel struct {
	Todos      []Todo
	db         *DB
	state      State
	viewpane   ViewPane
	editpane   EditPane
	pomodoro   tea.Model
	createpane CreatePane
}

func (aM AppModel) Init() tea.Cmd {
	return tea.Batch(aM.createpane.Init(), aM.editpane.Init())
}

func (aM *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch aM.state {
	case viewPane:
		updated, cmd := aM.viewpane.Update(msg)
		aM.viewpane = updated.(ViewPane)
		return aM, cmd
	case createPane:
		updated, cmd := aM.createpane.Update(msg)
		aM.createpane = updated.(CreatePane)
		return aM, cmd
	case editPane:
		updated, cmd := aM.editpane.Update(msg)
		aM.editpane = updated.(EditPane)
		return aM, cmd
	default:
		updated, cmd := aM.viewpane.Update(msg)
		aM.viewpane = updated.(ViewPane)
		return aM, cmd
	}
}

func (aM AppModel) View() string {
	switch aM.state {
	case viewPane:
		return appStyle.Height(termHeight).Width(termWidth).Render(aM.viewpane.View())
	case createPane:
		return appStyle.Height(termHeight).Width(termWidth).Render(aM.createpane.View())
	case editPane:
		return appStyle.Height(termHeight).Width(termWidth).Render(aM.editpane.View())
	default:
		return appStyle.Height(termHeight).Width(termWidth).Render(aM.createpane.View())
	}
}

func Render() {
	db, _ := NewDB()
	Todos := db.ReadTodo()
	app = &AppModel{db: db, Todos: Todos}
	app.db.CreateTodo(Todo{Title_: "Gae Man tries Golang", Description_: "Let's see if he's any good.", Status_: PENDING})
	app.db.CreateTodo(Todo{Title_: "Trying harder he is", Description_: "Won't make it he.", Status_: PENDING})
	InitViewPane()
	InitCreatePane()
	InitEditPane()
	if _, err := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseAllMotion()).Run(); err != nil {
		log.Fatal("the app frucked up")
	}
}
