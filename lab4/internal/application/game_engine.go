package application

import (
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/image/colornames"
)

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
		if int(points[0].X) >= g.GameSession.Grid.Width {
			points[0].X = 0
			points[1].X = int32(g.GameSession.Grid.Width - 1)
		}
		if int(points[0].X) < 0 {
			points[1].X = -int32(g.GameSession.Grid.Width - 1)
			points[0].X = int32(g.GameSession.Grid.Width - 1)
		}
		if int(points[0].Y) >= g.GameSession.Grid.Height {
			points[0].Y = 0
			points[1].Y = int32(g.GameSession.Grid.Height - 1)
		}
		if int(points[0].Y) < 0 {
			points[0].Y = int32(g.GameSession.Grid.Height - 1)
			points[1].Y = -int32(g.GameSession.Grid.Height - 1)
		}
		g.controllers[i].SetPoints(points)
	}
}

func (g *Game) checkFood() {
	for i, _ := range g.controllers {
		for k, food := range g.GameSession.State.Foods {
			points := g.controllers[i].GetPoints()
			head := points[0]
			curx := head.X
			cury := head.Y
			for j := 1; j < len(points); j++ {
				if (curx == food.X) && (cury == food.Y) {
					// here logic of growth
					g.controllers[i].GrowPlayer()
					g.GameSession.State.Foods = append(g.GameSession.State.Foods[:k], g.GameSession.State.Foods[k+1:]...)
					break
				}
				curx = curx + points[i].X
				cury = cury + points[i].Y
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
		err := g.discoverGame()
		if err != nil {
			return err
		}
		game, err := g.findGame()
		if err != nil {
			fmt.Println(err)
			break
		}
		g.JoinGame(game.Addr(), game.Msg.GetGameName(), game.Msg.GetCanJoin())
	case End:
		g.endGame()
	default:
		panic("unhandled default case")
	}
	return nil
}

func (g *Game) findGame() (AvailableGame, error) {
	g.availableGamesMutex.Lock()
	if g.availableGames == nil {
		g.availableGamesMutex.Unlock()
		return AvailableGame{}, errors.New("no available games")
	}
	for i, game := range g.availableGames {
		fmt.Printf("%d) %s, %v\n", i, game.Msg.GameName, game.Msg.CanJoin)
	}
	ind := -1
	_, err := fmt.Scan(&ind)
	if err != nil {
		g.availableGamesMutex.Unlock()
		return AvailableGame{}, err
	}
	if ind < 0 || ind >= len(g.availableGames) {
		g.availableGamesMutex.Unlock()
		return AvailableGame{}, errors.New("invalid game index")
	}
	game := g.availableGames[ind]
	g.availableGamesMutex.Unlock()
	return game, nil
}
