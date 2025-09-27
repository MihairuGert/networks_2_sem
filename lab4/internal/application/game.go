package application

import (
	"image"
	_ "image/jpeg"
	"snake-game/internal/application/ui"
	"snake-game/internal/domain"
	"snake-game/internal/infrastructure"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	screenWidthGlobal  = 640
	screenHeightGlobal = 480
)

type gameState int

const (
	Menu gameState = iota
	Connect
	Play
	End
)

type Game struct {
	Renderer    infrastructure.Renderer
	GameSession *GameSession
	Menu        *ui.Menu

	state gameState
}

type GameSession struct {
	Grid *domain.Grid
	domain.GameConfig
}

func (g *Game) Update() error {
	switch g.state {
	case Menu:
		g.Menu.Update()
	default:
		panic("unhandled default case")
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, strconv.FormatInt(int64(int(ebiten.ActualFPS())), 10))
	switch g.state {
	case Menu:
		g.Menu.Draw(screen)
	default:
		panic("unhandled default case")
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return screenWidthGlobal, screenHeightGlobal
}

func (g *Game) Init() {
	ebiten.SetWindowSize(screenWidthGlobal, screenHeightGlobal)
	ebiten.SetWindowTitle("Mihairu's Snake Game")
	_, icon, err := ebitenutil.NewImageFromFile("./textures/app_icon.jpeg")
	if err != nil {
		panic(err)
	}
	icons := []image.Image{icon}
	ebiten.SetWindowIcon(icons)
	g.state = Menu
	g.Menu = ui.NewMenu()

	//renderer := infrastructure.EbitRenderer{ScreenWidth: 640, ScreenHeight: 480}
	//g.GameSession = &GameSession{
	//	Grid:       domain.NewGrid(10, 10, 640, 480),
	//	GameConfig: domain.GameConfig{}}
	//g.Renderer = &renderer
	//g.Renderer.DrawGridImage(g.GameSession.Grid)

}

func (g *Game) Start() error {
	if err := ebiten.RunGame(g); err != nil {
		return err
	}
	return nil
}
