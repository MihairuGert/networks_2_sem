package application

import (
	"image/color"
	"snake-game/internal/application/ui"
	"snake-game/internal/domain"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/colornames"
)

func (g *Game) Draw(screen *ebiten.Image) {
	//ebitenutil.DebugPrint(screen, strconv.FormatInt(int64(int(ebiten.ActualFPS())), 10))
	screen.Fill(color.Black)
	switch g.state {
	case Menu:
		g.Menu.Draw(screen)
	case Play:
		g.Renderer.Draw(screen, g.GameSession)
		for _, c := range g.controllers {
			c.DrawPlayer(screen, g.GameSession.Grid)
		}
		g.drawFood(screen)
	case Connect:

	case End:
		g.finalMsg.Draw(screen)
	default:
		panic("unhandled default case")
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return screenWidthGlobal, screenHeightGlobal
}

func (g *Game) setupMenu() {
	g.Menu = ui.NewMenu()
	var err error

	g.Menu.AddMenuButton(screenWidthGlobal, screenHeightGlobal, g.handleNewGame)
	g.Menu.GetButton(0).NormalImage, _, err = ebitenutil.NewImageFromFile(texturesPath + "new_game.png")
	if err != nil {
		panic(err)
	}

	g.Menu.AddMenuButton(screenWidthGlobal, screenHeightGlobal, g.handleConnect)
	g.Menu.GetButton(1).NormalImage, _, err = ebitenutil.NewImageFromFile(texturesPath + "connect.png")
	if err != nil {
		panic(err)
	}

	g.Menu.AddMenuButton(screenWidthGlobal, screenHeightGlobal, g.handleExit)
	g.Menu.GetButton(2).NormalImage, _, err = ebitenutil.NewImageFromFile(texturesPath + "exit.png")
	if err != nil {
		panic(err)
	}
}

func (g *Game) handleNewGame() {
	g.state = Play

	renderer := ui.GameSessionRenderer{ScreenWidth: float32(screenWidthGlobal), ScreenHeight: float32(screenHeightGlobal)}
	g.GameSession = &domain.GameSession{
		Grid:   domain.NewGrid(20, 20, float32(screenWidthGlobal), float32(screenHeightGlobal)),
		Config: domain.GameConfig{}}
	g.Renderer = &renderer
	g.Renderer.SetGridImage(g.GameSession.Grid)

	controller := ui.Controller{}
	controller.SetPlayer(1, 1)
	g.addPlayer(&controller)

	g.lastFoodSpawnTime = time.Now()
	g.foodSpawnInt = time.Second * 3

	g.GameSession.BecomeMaster()
	g.goroutinePool.Go(g.startAnnouncement)
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
