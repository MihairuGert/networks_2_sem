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

	// Used in fancy exit window.
	shutdownTime  time.Time
	finalMsg      *ui.Text
	flickerInt    time.Duration
	lastFlickTime time.Time

	foodSpawnInt      time.Duration
	lastFoodSpawnTime time.Time
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

func (g *Game) checkBorders() {
	for i, _ := range g.controllers {
		points := g.controllers[i].GetPoints()
		if int(*points[0].X) >= g.GameSession.Grid.Width {
			*points[0].X = 0
			*points[1].X = int32(g.GameSession.Grid.Width - 1)
		}
		if int(*points[0].X) < 0 {
			*points[1].X = -int32(g.GameSession.Grid.Width - 1)
			*points[0].X = int32(g.GameSession.Grid.Width - 1)
		}
		if int(*points[0].Y) >= g.GameSession.Grid.Height {
			*points[0].Y = 0
			*points[1].Y = int32(g.GameSession.Grid.Height - 1)
		}
		if int(*points[0].Y) < 0 {
			*points[0].Y = int32(g.GameSession.Grid.Height - 1)
			*points[1].Y = -int32(g.GameSession.Grid.Height - 1)
		}
		g.controllers[i].SetPoints(points)
	}
}

func (g *Game) checkFood() {
	for i, _ := range g.controllers {
		for k, food := range g.GameSession.Foods {
			points := g.controllers[i].GetPoints()
			head := points[0]
			curx := *head.X
			cury := *head.Y
			for j := 1; j < len(points); j++ {
				if (curx == *food.X) && (cury == *food.Y) {
					// here logic of growth
					g.controllers[i].GrowPlayer()
					// careful! hz how slices work in go
					g.GameSession.Foods = append(g.GameSession.Foods[:k], g.GameSession.Foods[k+1:]...)
					break
				}
				curx = curx + *points[i].X
				cury = cury + *points[i].Y
			}
		}
	}
}

func (g *Game) addFood() {
	if time.Since(g.lastFoodSpawnTime) >= g.foodSpawnInt {
		g.lastFoodSpawnTime = time.Now()
		g.GameSession.GenerateFood(1)
	}
}

func (g *Game) Update() error {
	switch g.state {
	case Menu:
		g.Menu.Update()
	case Play:
		g.Renderer.Update()
		for i, _ := range g.controllers {
			g.controllers[i].Update()
		}
		g.checkBorders()
		g.checkFood()
		g.addFood()
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

func (g *Game) Init() {
	ebiten.SetWindowSize(screenWidthGlobal, screenHeightGlobal)
	ebiten.SetWindowTitle("Mihairu's Snake Game")
	_, icon, err := ebitenutil.NewImageFromFile(texturesPath + "app_icon.jpeg")
	if err != nil {
		panic(err)
	}
	icons := []image.Image{icon}
	ebiten.SetWindowIcon(icons)

	g.lastFoodSpawnTime = time.Now()
	g.foodSpawnInt = time.Second * 3

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
		Grid:       domain.NewGrid(20, 20, float32(screenWidthGlobal), float32(screenHeightGlobal)),
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

func (g *Game) drawFood(screen *ebiten.Image) {
	for _, Food := range g.GameSession.Foods {
		rectImage := ebiten.NewImage(int(g.GameSession.Grid.RectWidth), int(g.GameSession.Grid.RectHeight))
		rectImage.Fill(colornames.Darkred)

		curX := float64(*Food.X) * float64(g.GameSession.Grid.RectWidth)
		curY := float64(*Food.Y) * float64(g.GameSession.Grid.RectHeight)

		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(curX, curY)
		screen.DrawImage(rectImage, opts)
	}
}
