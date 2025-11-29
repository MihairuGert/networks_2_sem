package ui

import (
	"fmt"
	"snake-game/internal/domain"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/colornames"
)

type GameSessionRenderer struct {
	gridImage    *ebiten.Image
	ScreenWidth  float32
	ScreenHeight float32
	PlayerList   *PlayerList
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

func (r *GameSessionRenderer) Update(players []*domain.GamePlayer) {
	if r.PlayerList != nil {
		r.PlayerList.Update(players)
	}
}

func (r *GameSessionRenderer) DrawGrid(screen *ebiten.Image) {
	screen.Fill(colornames.Black)
	screen.DrawImage(r.gridImage, &ebiten.DrawImageOptions{})
}

func (r *GameSessionRenderer) DrawPlayerList(screen *ebiten.Image) {
	if r.PlayerList != nil {
		r.PlayerList.Draw(screen)
	}
}

type PlayerList struct {
	playerTexts []*Text
	x, y        float64
	step        float64
}

func NewPlayerList(x, y, step float64) *PlayerList {
	return &PlayerList{
		x:    x,
		y:    y,
		step: step,
	}
}

func (pl *PlayerList) Update(players []*domain.GamePlayer) {
	pl.playerTexts = nil

	for i, player := range players {
		text := NewText("", 14, pl.x, pl.y+float64(i)*pl.step)
		text.SetText(pl.formatPlayer(player))
		text.SetColor(colornames.White)
		pl.playerTexts = append(pl.playerTexts, text)
	}
}

func (pl *PlayerList) formatPlayer(player *domain.GamePlayer) string {
	role := "?"
	switch player.Role {
	case domain.NodeRole_NORMAL:
		role = "P"
	case domain.NodeRole_MASTER:
		role = "M"
	case domain.NodeRole_DEPUTY:
		role = "D"
	case domain.NodeRole_VIEWER:
		role = "V"
	}
	return fmt.Sprintf("%s[%s] - %d", player.Name, role, player.Score)
}

func (pl *PlayerList) Draw(screen *ebiten.Image) {
	for _, text := range pl.playerTexts {
		text.Draw(screen)
	}
}
