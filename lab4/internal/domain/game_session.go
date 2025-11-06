package domain

import "math/rand"

type GameSession struct {
	Node Node
	Grid *Grid

	Config  GameConfig
	Players GamePlayers
	State   GameState

	nextPlayerId    int
	currentStateNum int
}

func (gs *GameSession) IncrementStateNum() {
	gs.currentStateNum++
}

func (gs *GameSession) CurrentStateNum() int {
	return gs.currentStateNum
}

// GetFreePlayerId guarantees to return a free id.
func (gs *GameSession) GetFreePlayerId() int {
	temp := gs.nextPlayerId
	gs.nextPlayerId++
	return temp
}

func (gs *GameSession) BecomeMaster() {
	gs.Node.role = NodeRole_MASTER
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
			gs.State.Foods = append(gs.State.Foods, &food)
			coordsX[x] = true
			coordsY[y] = true
		}
	}
}
