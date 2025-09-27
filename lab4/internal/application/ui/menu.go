package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type Menu struct {
	buttons []*Button
}

func NewMenu() *Menu {
	buttons := []*Button{
		NewButton(10, 10, 100, 100, "hui", func() { fmt.Print("JOPA") }),
	}
	return &Menu{
		buttons: buttons,
	}
}

func (m *Menu) addButton(b *Button) {
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
