package application

import (
	"context"
	"snake-game/internal/application/network"
	"snake-game/internal/application/ui"
	"snake-game/internal/domain"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"
	"golang.org/x/sync/errgroup"
)

func (g *Game) handleNewGame() {
	g.state = Play

	g.startGame()

	g.startNetwork()
}

func (g *Game) startGame() {
	renderer := ui.GameSessionRenderer{ScreenWidth: float32(screenWidthGlobal), ScreenHeight: float32(screenHeightGlobal)}

	config, err := ui.ParseConfig("conf.yaml")
	if err != nil {
		panic(err)
	}
	g.GameSession = domain.NewGameSession(config, float32(screenWidthGlobal), float32(screenHeightGlobal))
	g.Renderer = &renderer
	g.Renderer.SetGridImage(g.GameSession.Grid)

	g.GameSession.BecomeMaster()

	g.controllers = make(map[int]domain.Controller)
	controller := ui.Controller{}
	controller.SetPlayer(1, 1)
	g.addPlayer(&controller)

	g.lastFoodSpawnTime = time.Now()
	g.foodSpawnInt = time.Second * 3
}

func (g *Game) startNetwork() {
	// todo move it to connect!
	//g.availableGames = make(map[string]AvailableGame)
	g.networkManager = network.NewNetworkManager()
	g.goroutinePool, _ = errgroup.WithContext(context.Background())
	g.goroutinePool.Go(g.startAnnouncement)
	g.startListening()
}

func (g *Game) handleConnect() {
	g.state = Connect
}

func (g *Game) handleExit() {
	g.state = End
	g.shutdownTime = time.Now()
	g.lastFlickTime = time.Now()
	g.flickerInt = 25 * time.Millisecond
	g.finalMsg = ui.NewText("", 24, 100, 100)
}

func (g *Game) drawFood(screen *ebiten.Image) {
	for _, Food := range g.GameSession.State.Foods {
		rectImage := ebiten.NewImage(int(g.GameSession.Grid.RectWidth), int(g.GameSession.Grid.RectHeight))
		rectImage.Fill(colornames.Darkred)

		curX := float64(Food.X) * float64(g.GameSession.Grid.RectWidth)
		curY := float64(Food.Y) * float64(g.GameSession.Grid.RectHeight)

		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(curX, curY)
		screen.DrawImage(rectImage, opts)
	}
}
