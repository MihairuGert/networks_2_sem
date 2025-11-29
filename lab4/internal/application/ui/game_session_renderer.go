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

func (r *GameSessionRenderer) Update() {
	r.PlayerList.Update()
}

func (r *GameSessionRenderer) DrawGrid(screen *ebiten.Image) {
	screen.Fill(colornames.Black)
	screen.DrawImage(r.gridImage, &ebiten.DrawImageOptions{})
}

func (r *GameSessionRenderer) DrawPlayerList(screen *ebiten.Image) {
	r.PlayerList.Draw(screen)
}

type PlayerList struct {
	players []*PlayerEntry
	x, y    float64
	step    float64
}

type PlayerEntry struct {
	name  string
	role  string
	score int
	text  *Text
}

func NewPlayerList(x, y, step float64) *PlayerList {
	return &PlayerList{
		x:    x,
		y:    y,
		step: step,
	}
}

func (r *GameSessionRenderer) AddPlayer(name, role string, score int) {
	entry := &PlayerEntry{
		name:  name,
		role:  role,
		score: score,
		text:  NewText("", 16, r.PlayerList.x, r.PlayerList.y+float64(len(r.PlayerList.players))*r.PlayerList.step),
	}
	entry.text.SetText(entry.format())
	entry.text.SetColor(colornames.White)
	r.PlayerList.players = append(r.PlayerList.players, entry)
}

func (r *GameSessionRenderer) RemovePlayer(name string) {
	for i, player := range r.PlayerList.players {
		if player.name == name {
			r.PlayerList.players = append(r.PlayerList.players[:i], r.PlayerList.players[i+1:]...)
			r.PlayerList.updatePositions()
			return
		}
	}
}

func (r *GameSessionRenderer) UpdatePlayerScore(name string, score int) {
	for _, player := range r.PlayerList.players {
		if player.name == name {
			player.score = score
			player.text.SetText(player.format())
			return
		}
	}
}

func (r *GameSessionRenderer) UpdatePlayerRole(name string, role string) {
	for _, player := range r.PlayerList.players {
		if player.name == name {
			player.role = role
			player.text.SetText(player.format())
			return
		}
	}
}

func (player *PlayerEntry) format() string {
	return fmt.Sprintf("%s[%s]-%d", player.name, player.role, player.score)
}

func (pl *PlayerList) Update() {
	for _, player := range pl.players {
		player.text.SetColor(colornames.White)
	}
}

func (pl *PlayerList) Draw(screen *ebiten.Image) {
	for _, player := range pl.players {
		player.text.Draw(screen)
	}
}

func (pl *PlayerList) updatePositions() {
	for i, player := range pl.players {
		player.text.SetPosition(pl.x, pl.y+float64(i)*pl.step)
	}
}

func (pl *PlayerList) Clear() {
	pl.players = []*PlayerEntry{}
}
