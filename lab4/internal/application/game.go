package application

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"os"
	"snake-game/internal/application/ui"
	"snake-game/internal/domain"
	"snake-game/internal/infrastructure"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	screenWidthGlobal  = 640
	screenHeightGlobal = 480
)

const texturesPath = "./textures/"

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

	shutdownTime time.Time
}

type GameSession struct {
	Grid *domain.Grid
	domain.GameConfig
}

func (g *Game) endGame() {
	elapsed := time.Since(g.shutdownTime)
	if elapsed >= 1*time.Second {
		os.Exit(0)
	}
}

func (g *Game) Update() error {
	switch g.state {
	case Menu:
		g.Menu.Update()
	case End:
		g.endGame()
	default:
		panic("unhandled default case")
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	//ebitenutil.DebugPrint(screen, strconv.FormatInt(int64(int(ebiten.ActualFPS())), 10))
	screen.Fill(color.Black)
	switch g.state {
	case Menu:
		g.Menu.Draw(screen)
	case End:
		ebitenutil.DebugPrint(screen, "That's... the end")
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
	_, icon, err := ebitenutil.NewImageFromFile(texturesPath + "app_icon.jpeg")
	if err != nil {
		panic(err)
	}
	icons := []image.Image{icon}
	ebiten.SetWindowIcon(icons)

	g.state = Menu
	g.setupMenu()

	//renderer := infrastructure.EbitRenderer{ScreenWidth: 640, ScreenHeight: 480}
	//g.GameSession = &GameSession{
	//	Grid:       domain.NewGrid(10, 10, 640, 480),
	//	GameConfig: domain.GameConfig{}}
	//g.Renderer = &renderer
	//g.Renderer.DrawGridImage(g.GameSession.Grid)

}

func (g *Game) setupMenu() {
	g.Menu = ui.NewMenu()
	var err error

	g.Menu.AddMenuButton(screenWidthGlobal, screenHeightGlobal, func() { fmt.Print("1") })
	g.Menu.GetButton(0).NormalImage, _, err = ebitenutil.NewImageFromFile(texturesPath + "new_game.png")
	if err != nil {
		panic(err)
	}

	g.Menu.AddMenuButton(screenWidthGlobal, screenHeightGlobal, func() { fmt.Print("2") })
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

func (g *Game) Start() error {
	if err := ebiten.RunGame(g); err != nil {
		return err
	}
	return nil
}

func (g *Game) handleExit() {
	g.state = End
	g.shutdownTime = time.Now()
}
