package models

import (
	tea "github.com/charmbracelet/bubbletea"
)

type DefaultModel interface {
	Value() string
	Update(tea.Msg) (DefaultModel, tea.Cmd)
	View() string
	Type() string
	SetType(t string)
	Title() string
	SetTitle(string)
}

type BlurAble interface {
	Blur() tea.Msg
}
