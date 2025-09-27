package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
)

type Button struct {
	Rect    image.Rectangle
	Text    string
	OnClick func()
	hovered bool
	pressed bool
}

func NewButton(x, y, width, height int, text string, onClick func()) *Button {
	return &Button{
		Rect:    image.Rect(x, y, x+width, y+height),
		Text:    text,
		OnClick: onClick,
	}
}

func (b *Button) Update() {
	x, y := ebiten.CursorPosition()

	b.hovered = image.Pt(x, y).In(b.Rect)

	if b.hovered && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		b.pressed = true
	}

	if b.pressed && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if b.hovered && b.OnClick != nil {
			b.OnClick()
		}
		b.pressed = false
	}
}

func (b *Button) Draw(screen *ebiten.Image) {
	img := ebiten.NewImage(b.Rect.Dx(), b.Rect.Dy())
	img.Fill(colornames.Crimson)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.Rect.Min.X), float64(b.Rect.Min.Y))
	screen.DrawImage(img, op)

}
