package domain

import "math/rand"

type Controller interface {
	Move(Direction)
	Kill()
}

type GameSession struct {
	Grid *Grid

	GameConfig
	GamePlayers
	GameState
}

// GenerateFood it is not guaranteed to generate all asked count so far.
func (gs *GameSession) GenerateFood(count int) {
	coordsX := make(map[int32]bool)
	coordsY := make(map[int32]bool)
	for i := 0; i < count; i++ {
		x := int32(rand.Intn(gs.Grid.Width))
		y := int32(rand.Intn(gs.Grid.Height))
		food := GameState_Coord{X: x, Y: y}
		_, okX := coordsX[x]
		_, okY := coordsY[y]
		if !okX && !okY {
			gs.Foods = append(gs.Foods, &food)
			coordsX[x] = true
			coordsY[y] = true
		}
	}
}
