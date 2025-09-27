package ui

import (
	"bytes"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Text struct {
	text  string
	size  float64
	X     float64
	Y     float64
	color color.Color
}

var (
	mplusFaceSource *text.GoTextFaceSource
)

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.PressStart2P_ttf))
	if err != nil {
		log.Fatal(err)
	}
	mplusFaceSource = s
}

func NewText(text string, size, x, y float64) *Text {
	return &Text{
		text:  text,
		size:  size,
		X:     x,
		Y:     y,
		color: color.White,
	}
}

func (t *Text) Draw(screen *ebiten.Image) {
	face := &text.GoTextFace{
		Source: mplusFaceSource,
		Size:   t.size,
	}

	op := &text.DrawOptions{}
	op.GeoM.Translate(t.X, t.Y)
	op.ColorScale.ScaleWithColor(t.color)
	text.Draw(screen, t.text, face, op)
}

func (t *Text) SetText(newText string) {
	t.text = newText
}

func (t *Text) SetPosition(x, y float64) {
	t.X = x
	t.Y = y
}

func (t *Text) SetColor(c color.Color) {
	t.color = c
}
