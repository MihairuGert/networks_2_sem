package infrastructure

import (
	"image/color"
	"snake-game/internal/domain"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/colornames"
)

type Renderer interface {
	DrawGridImage(grid *domain.Grid)
	GetGridImage() *ebiten.Image
}

type EbitRenderer struct {
	gridImage    *ebiten.Image
	ScreenWidth  float32
	ScreenHeight float32
}

func (r *EbitRenderer) GetGridImage() *ebiten.Image {
	return r.gridImage
}

func (r *EbitRenderer) DrawGridImage(grid *domain.Grid) {
	image := ebiten.NewImage(int(r.ScreenWidth), int(r.ScreenHeight))
	image.Fill(color.White)
	for i := float32(0); i < r.ScreenWidth; i += grid.RectWidth {
		vector.StrokeLine(image, i, 0, i, r.ScreenHeight, 1, colornames.Burlywood, true)
	}
	for i := float32(0); i < r.ScreenHeight; i += grid.RectHeight {
		vector.StrokeLine(image, 0, i, r.ScreenWidth, i, 1, colornames.Burlywood, true)
	}
	r.gridImage = image
}
