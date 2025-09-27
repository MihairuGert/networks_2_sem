package ui

import (
	"image/color"
	"snake-game/internal/domain"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/colornames"
)

type GameSessionRenderer struct {
	gridImage    *ebiten.Image
	ScreenWidth  float32
	ScreenHeight float32
}

func (r *GameSessionRenderer) GetGridImage() *ebiten.Image {
	return r.gridImage
}

func (r *GameSessionRenderer) SetGridImage(grid *domain.Grid) {
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

func (r *GameSessionRenderer) Update() {

}

func (r *GameSessionRenderer) Draw(screen *ebiten.Image, session *domain.GameSession) {
	screen.DrawImage(r.gridImage, &ebiten.DrawImageOptions{})
}
