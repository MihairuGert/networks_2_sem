package domain

import (
	"math/rand"
	"time"
)

type GameSession struct {
	Node Node
	Grid *Grid

	Players []*PlayerWrapper
	Config  *GameConfig
	State   *GameState

	myID int

	nextPlayerId    int
	currentStateNum int

	LastIterationTime time.Time
}

func (gs *GameSession) MyID() int {
	return gs.myID
}

func (gs *GameSession) SetMyID(myID int) {
	gs.myID = myID
}

func NewGameSession(config *GameConfig, screenWidth, screenHeight float32) *GameSession {
	grid := NewGrid(int(config.Width), int(config.Height), screenWidth, screenHeight)
	gameState := &GameState{Snakes: make([]*GameState_Snake, 0), Players: nil, Foods: make([]*GameState_Coord, 0)}
	return &GameSession{
		Grid:   grid,
		Config: config,
		State:  gameState}
}

func (gs *GameSession) StateDelayMs() int32 {
	return gs.Config.StateDelayMs
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

func (gs *GameSession) BecomeNormal() {
	gs.Node.role = NodeRole_NORMAL
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
