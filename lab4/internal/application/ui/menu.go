package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Menu struct {
	buttons []*Button
	offset  int
	step    int
}

func NewMenu() *Menu {
	step := 25
	return &Menu{offset: step * 4, step: step}
}

func (m *Menu) GetButton(ind int) *Button {
	return m.buttons[ind]
}

func (m *Menu) AddButton(b *Button) {
	m.buttons = append(m.buttons, b)
}

func (m *Menu) Update() {
	for _, b := range m.buttons {
		b.Update()
	}
}

func (m *Menu) Draw(screen *ebiten.Image) {
	for _, b := range m.buttons {
		b.Draw(screen)
	}
}

func (m *Menu) AddMenuButton(screenWidth, screenHeight int, f func()) {
	width := screenWidth / 4
	height := screenHeight / 7
	defer func() { m.offset += height + m.step }()
	m.AddButton(NewButton((screenWidth-width)/2, m.offset, width, height, f))
}
