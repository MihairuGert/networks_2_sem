package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
)

type Button struct {
	Rect         image.Rectangle
	NormalImage  *ebiten.Image
	HoverImage   *ebiten.Image
	PressedImage *ebiten.Image
	OnClick      func()
	hovered      bool
	pressed      bool
}

func NewButton(x, y, width, height int, onClick func()) *Button {
	return &Button{
		Rect:    image.Rect(x, y, x+width, y+height),
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
	var img *ebiten.Image
	switch {
	case b.pressed && b.PressedImage != nil:
		img = b.PressedImage
	case b.hovered && b.HoverImage != nil:
		img = b.HoverImage
	default:
		if b.NormalImage != nil {
			img = b.NormalImage
		} else {
			img = ebiten.NewImage(b.Rect.Dx(), b.Rect.Dy())
			img.Fill(colornames.Crimson)
		}
	}

	if img != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(b.Rect.Dx())/float64(img.Bounds().Dx()), float64(b.Rect.Dy())/float64(img.Bounds().Dy()))
		op.GeoM.Translate(float64(b.Rect.Min.X), float64(b.Rect.Min.Y))
		screen.DrawImage(img, op)
	}

}
