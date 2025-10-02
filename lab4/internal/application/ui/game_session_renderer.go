package ui

import (
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
	xSize := grid.RectWidth * float32(grid.Width)
	ySize := grid.RectHeight * float32(grid.Height)
	for i := float32(0); i <= float32(grid.Width); i += 1 {
		vector.StrokeLine(image, i*grid.RectWidth, 0, i*grid.RectWidth, ySize, 1, colornames.Crimson, true)
	}
	for i := float32(0); i <= float32(grid.Height); i += 1 {
		vector.StrokeLine(image, 0, i*grid.RectHeight, xSize, i*grid.RectHeight, 1, colornames.Crimson, true)
	}
	r.gridImage = image
}

func (r *GameSessionRenderer) Update() {

}

func (r *GameSessionRenderer) Draw(screen *ebiten.Image, session *domain.GameSession) {
	screen.Fill(colornames.Black)
	screen.DrawImage(r.gridImage, &ebiten.DrawImageOptions{})
}
