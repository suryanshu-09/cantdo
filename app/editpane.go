package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type EditPane struct {
	form *huh.Form
}

func (eP EditPane) Init() tea.Cmd { return nil }

func (eP EditPane) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		termHeight = msg.Height
		termWidth = msg.Width
	}
	updated, cmd := eP.form.Update(msg)
	eP.form = updated.(*huh.Form)
	return eP, cmd
}

func (eP EditPane) View() string {
	return paneStyle.Render(eP.form.View())
}
