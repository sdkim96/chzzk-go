package main

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type Model struct {
	title string
	state *state
}

type state struct {
	update int
	status int
	prompt textinput.Model
}

func (s *state) isSystemMode() bool {
	if s.status == 0 {
		return true
	}
	return false
}

func NewModel() Model {
	return Model{
		title: "Chzzk Viewer",
		state: &state{prompt: textinput.New()},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.state.update++
	kpMsg, ok := msg.(tea.KeyPressMsg)
	if ok && isEOF(kpMsg) {
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) View() tea.View {
	view := fmt.Sprintf("%s\n\n%d", m.title, m.state.update)
	return tea.NewView(view)
}

func main() {
	if _, err := tea.NewProgram(NewModel()).Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
