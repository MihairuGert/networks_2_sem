package infrastructure

import (
	"image/color"
	"snake-game/internal/domain"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/colornames"
)

type Renderer struct {
	ScreenWidth, ScreenHeight float32
}

func (r Renderer) GetGridImage(grid *domain.Grid) *ebiten.Image {
	image := &ebiten.Image{}
	image.Fill(color.White)
	for i := float32(0); i < r.ScreenWidth; i += grid.RectWidth {
		vector.StrokeLine(image, i, 0, r.ScreenHeight, i, 1, colornames.Burlywood, true)
	}
	for i := float32(0); i < r.ScreenHeight; i += grid.RectHeight {
		vector.StrokeLine(image, 0, i, r.ScreenWidth, i, 1, colornames.Burlywood, true)
	}
	return image
}
