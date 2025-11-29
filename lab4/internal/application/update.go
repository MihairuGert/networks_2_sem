package application

import (
	"errors"
	"fmt"
	"os"
	"snake-game/internal/domain"
	"time"

	"golang.org/x/image/colornames"
)

var Recheck = errors.New("recheck")
var Exit = errors.New("exit")

func (g *Game) endApplication() {
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
		if g.GameSession.Players[i].Snake == nil {
			continue
		}
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
	for i := range g.GameSession.Players {
		if g.GameSession.Players[i].Snake == nil {
			continue
		}

		points := g.GameSession.Players[i].GetPoints()
		if len(points) == 0 {
			continue
		}

		head := points[0]

		var remainingFood []*domain.GameState_Coord
		foodEaten := false

		for _, food := range g.GameSession.State.Foods {
			if head.X == food.X && head.Y == food.Y && !foodEaten {
				g.GameSession.Players[i].Grow()
				foodEaten = true
			} else {
				remainingFood = append(remainingFood, food)
			}
		}

		g.GameSession.State.Foods = remainingFood
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
			return nil
		}
		g.controller.Update()
		g.Renderer.ExitButton.Update()
		if g.shouldStop {
			return nil
		}
		switch g.GameSession.Node.Role() {
		case domain.NodeRole_MASTER:
			if time.Since(g.GameSession.LastIterationTime) >= time.Duration(g.GameSession.StateDelayMs())*time.Millisecond {
				g.GameSession.LastIterationTime = time.Now()
				g.computeNextIteration()
				g.setState()
				g.Renderer.Update(g.GameSession.State.Players.Players)
				g.sendState()
			}
		case domain.NodeRole_DEPUTY:
			g.sendSteer()
		case domain.NodeRole_NORMAL:
			g.sendSteer()
		case domain.NodeRole_VIEWER:
		}
	case End:
		g.endApplication()
	default:
		panic("unhandled default case")
	}
	return nil
}

func (g *Game) computeNextIteration() {
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
		fmt.Printf("No available games. Write e to exit, any other key to recheck...\n")
		ind := ""
		fmt.Scan(&ind)
		if ind == "e" {
			g.handleExitGame()
			return AvailableGame{}, Exit
		}
		return AvailableGame{}, errors.New("no available games")
	}

	availableGames := make([]AvailableGame, 0)
	for _, game := range g.availableGames {
		availableGames = append(availableGames, game)
	}
	g.availableGamesMutex.Unlock()
	if len(availableGames) == 0 {
		fmt.Printf("No available games. Write e to exit, any other key to recheck...\n")
		ind := ""
		fmt.Scan(&ind)
		if ind == "e" {
			g.handleExitGame()
			return AvailableGame{}, Exit
		}
		return AvailableGame{}, errors.New("no available games")
	}

	for i, game := range availableGames {
		fmt.Printf("%d) %s, %v\n", i, game.Msg.GameName, game.Msg.CanJoin)
	}
	fmt.Printf("%d) Recheck\n", len(availableGames))
	ind := -1
	_, err := fmt.Scan(&ind)
	if err != nil {
		return AvailableGame{}, err
	}
	if ind == len(availableGames) {
		return AvailableGame{}, Recheck
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
