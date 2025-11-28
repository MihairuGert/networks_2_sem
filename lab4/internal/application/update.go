package application

import (
	"errors"
	"fmt"
	"os"
	"snake-game/internal/domain"
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
	for i, _ := range g.GameSession.Players {
		points := g.GameSession.Players[i].GetPoints()
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
		g.GameSession.Players[i].SetPoints(points)
	}
}

func (g *Game) checkFood() {
	for i, _ := range g.GameSession.Players {
		for k, food := range g.GameSession.State.Foods {
			points := g.GameSession.Players[i].GetPoints()
			head := points[0]
			curx := head.X
			cury := head.Y
			for j := 1; j < len(points); j++ {
				if (curx == food.X) && (cury == food.Y) {
					// here logic of growth
					g.GameSession.Players[i].Grow()
					g.GameSession.State.Foods = append(g.GameSession.State.Foods[:k], g.GameSession.State.Foods[k+1:]...)
					break
				}
				curx = curx + points[j].X
				cury = cury + points[j].Y
			}
		}
	}
}

func (g *Game) addFood() {
	if time.Since(g.lastFoodSpawnTime) >= g.foodSpawnInt {
		g.lastFoodSpawnTime = time.Now()
		g.GameSession.GenerateFood()
	}
}

func (g *Game) Update() error {
	switch g.state {
	case Menu:
		g.Menu.Update()
	case Play:
		err := g.handleIncomingMessages()
		if err != nil {
			fmt.Println(err)
		}
		g.controller.Update()
		switch g.GameSession.Node.Role() {
		case domain.NodeRole_MASTER:
			if time.Since(g.GameSession.LastIterationTime) >= time.Duration(g.GameSession.StateDelayMs())*time.Millisecond {
				g.GameSession.LastIterationTime = time.Now()
				g.computeNextIteration()
				g.setState()
				g.sendState()
			}
		case domain.NodeRole_DEPUTY:
			g.sendSteer()
		case domain.NodeRole_NORMAL:
			g.sendSteer()
		case domain.NodeRole_VIEWER:
			g.sendSteer()
		}
	case End:
		g.endGame()
	default:
		panic("unhandled default case")
	}
	return nil
}

func (g *Game) computeNextIteration() {
	g.Renderer.Update()
	g.moveControllers()
	g.checkBorders()
	g.checkPlayerCollision()
	g.checkFood()
	g.addFood()
	g.GameSession.IncrementStateNum()
}

func (g *Game) findGame() (AvailableGame, error) {
	g.availableGamesMutex.Lock()
	if g.availableGames == nil {
		g.availableGamesMutex.Unlock()
		return AvailableGame{}, errors.New("no available games")
	}

	availableGames := make([]AvailableGame, 0)
	for _, game := range g.availableGames {
		availableGames = append(availableGames, game)
	}
	g.availableGamesMutex.Unlock()
	if len(availableGames) == 0 {
		return AvailableGame{}, errors.New("no available games")
	}

	for i, game := range availableGames {
		fmt.Printf("%d) %s, %v\n", i, game.Msg.GameName, game.Msg.CanJoin)
	}
	ind := -1
	_, err := fmt.Scan(&ind)
	if err != nil {
		return AvailableGame{}, err
	}
	if ind < 0 || ind >= len(g.availableGames) {
		return AvailableGame{}, errors.New("invalid game index")
	}
	game := availableGames[ind]
	return game, nil
}

func (g *Game) moveControllers() {
	for i, _ := range g.GameSession.Players {
		g.GameSession.Players[i].Move()
	}
}

func (g *Game) checkPlayerCollision() {
	g.GameSession.CheckCollisions()
}
