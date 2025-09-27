package main

import (
	"log"
	"snake-game/internal/domain"
	"snake-game/internal/infrastructure"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	Renderer    infrastructure.Renderer
	gameSession *GameSession
}

type GameSession struct {
	Grid *domain.Grid
	domain.GameConfig
}

func (g *Game) Update() error {
	// Write your game's logical update.
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, strconv.FormatInt(int64(int(ebiten.ActualFPS())), 10))
	screen.DrawImage(g.Renderer.GetGridImage(), &ebiten.DrawImageOptions{})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}

func main() {
	renderer := infrastructure.EbitRenderer{ScreenWidth: 640, ScreenHeight: 480}
	game := &Game{gameSession: &GameSession{
		Grid:       domain.NewGrid(10, 10, 640, 480),
		GameConfig: domain.GameConfig{},
	}, Renderer: &renderer}
	game.Renderer.DrawGridImage(game.gameSession.Grid)
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Your game's title")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
