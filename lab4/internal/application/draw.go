package application

import (
	"image/color"
	"snake-game/internal/application/ui"
	"snake-game/internal/domain"

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
		if g.GameSession == nil || g.GameSession.State == nil {
			return
		}
		g.Renderer.DrawGrid(screen)
		g.Renderer.DrawPlayerList(screen)
		g.drawSnakes(screen)
		g.drawFood(screen)
	case End:
		g.finalMsg.Draw(screen)
	default:
		panic("unhandled default case")
	}
}

func (g *Game) drawSnakes(screen *ebiten.Image) {
	if g.GameSession.State.Snakes == nil {
		return
	}
	for _, snake := range g.GameSession.State.Snakes {
		drawSnake(screen, snake, g.GameSession.Grid)
	}
}

func drawSnake(screen *ebiten.Image, snake *domain.GameState_Snake, grid *domain.Grid) {
	rectImage := ebiten.NewImage(int(grid.RectWidth), int(grid.RectHeight))
	rectImage.Fill(color.RGBA{R: 255, G: 0, B: 0, A: 255})

	curX := float64(snake.Points[0].X) * float64(grid.RectWidth)
	curY := float64(snake.Points[0].Y) * float64(grid.RectHeight)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(curX, curY)
	screen.DrawImage(rectImage, opts)

	for i := 1; i < len(snake.Points); i++ {
		curX = curX + float64(grid.RectWidth)*float64(snake.Points[i].X)
		curY = curY + float64(grid.RectHeight)*float64(snake.Points[i].Y)

		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(curX, curY)
		screen.DrawImage(rectImage, opts)
	}
}

func (g *Game) drawFood(screen *ebiten.Image) {
	if g.GameSession.State.Foods == nil {
		return
	}
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
