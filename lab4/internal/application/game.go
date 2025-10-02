package application

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"os"
	"snake-game/internal/application/ui"
	"snake-game/internal/domain"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/colornames"
)

var (
	screenWidthGlobal  = 960
	screenHeightGlobal = 720
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
	Renderer    *ui.GameSessionRenderer
	GameSession *domain.GameSession
	Menu        *ui.Menu
	controllers []ui.Controller

	state gameState

	shutdownTime  time.Time
	finalMsg      *ui.Text
	flickerInt    time.Duration
	lastFlickTime time.Time
}

func (g *Game) endGame() {
	elapsed := time.Since(g.shutdownTime)
	g.finalMsg.SetText("Goodbye!!")
	g.finalMsg.SetColor(colornames.White)
	if time.Since(g.lastFlickTime) >= g.flickerInt {
		g.finalMsg.SetText("die.")
		g.finalMsg.SetColor(colornames.Crimson)
		g.lastFlickTime = time.Now()
	}
	if elapsed >= 3*time.Second {
		os.Exit(0)
	}
}

func (g *Game) Update() error {
	switch g.state {
	case Menu:
		g.Menu.Update()
	case Play:
		g.Renderer.Update()
		for _, c := range g.controllers {
			c.Update()
		}
	case Connect:

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
	case Play:
		g.Renderer.Draw(screen, g.GameSession)
		for _, c := range g.controllers {
			c.DrawPlayer(screen, g.GameSession.Grid)
		}
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
}

func (g *Game) addPlayer(c ui.Controller) {
	g.controllers = append(g.controllers, c)
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

func (g *Game) Start() error {
	if err := ebiten.RunGame(g); err != nil {
		return err
	}
	return nil
}

func (g *Game) handleNewGame() {
	g.state = Play
	renderer := ui.GameSessionRenderer{ScreenWidth: float32(screenWidthGlobal), ScreenHeight: float32(screenHeightGlobal)}
	g.GameSession = &domain.GameSession{
		Grid:       domain.NewGrid(10, 10, float32(screenWidthGlobal), float32(screenHeightGlobal)),
		GameConfig: domain.GameConfig{}}
	g.Renderer = &renderer
	g.Renderer.SetGridImage(g.GameSession.Grid)

	controller := ui.Controller{}
	controller.SetPlayer(1, 1)
	g.addPlayer(controller)
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
