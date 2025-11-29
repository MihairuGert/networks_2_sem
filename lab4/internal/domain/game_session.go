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

	Me *PlayerWrapper

	myID int32

	nextPlayerId    int32
	currentStateNum int

	LastIterationTime time.Time
}

func (gs *GameSession) MyID() int32 {
	return gs.myID
}

func (gs *GameSession) SetMyID(myID int32) {
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
func (gs *GameSession) GetFreePlayerId() int32 {
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

func (gs *GameSession) BecomeViewer() {
	gs.Node.role = NodeRole_VIEWER
}

func (gs *GameSession) GenerateFood() {
	requiredFood := gs.Config.FoodStatic + gs.countAliveSnakes()

	currentFood := int32(len(gs.State.Foods))

	if currentFood < requiredFood {
		gs.addFood(int(requiredFood - currentFood))
	}
}

func (gs *GameSession) countAliveSnakes() int32 {
	count := int32(0)
	for _, player := range gs.Players {
		if player != nil && player.Snake != nil && player.Snake.State == GameState_Snake_ALIVE {
			count++
		}
	}
	return count
}

func (gs *GameSession) addFood(count int) {
	if count <= 0 {
		return
	}

	freeCells := gs.getFreeCells()

	rand.Shuffle(len(freeCells), func(i, j int) {
		freeCells[i], freeCells[j] = freeCells[j], freeCells[i]
	})

	toAdd := count
	if toAdd > len(freeCells) {
		toAdd = len(freeCells)
	}

	for i := 0; i < toAdd; i++ {
		gs.State.Foods = append(gs.State.Foods, freeCells[i])
	}
}

func (gs *GameSession) getOccupiedCells() map[string]bool {
	occupied := make(map[string]bool)

	for _, player := range gs.Players {
		if player == nil || player.Snake == nil {
			continue
		}

		snakeCoords := gs.getSnakeAbsoluteCoordinates(player.Snake)
		for _, coord := range snakeCoords {
			occupied[gs.coordKey(coord.X, coord.Y)] = true
		}
	}

	for _, food := range gs.State.Foods {
		occupied[gs.coordKey(food.X, food.Y)] = true
	}

	return occupied
}

func (gs *GameSession) getFreeCells() []*GameState_Coord {
	occupied := gs.getOccupiedCells()
	var freeCells []*GameState_Coord

	for x := int32(0); x < gs.Config.Width; x++ {
		for y := int32(0); y < gs.Config.Height; y++ {
			if !occupied[gs.coordKey(x, y)] {
				freeCells = append(freeCells, &GameState_Coord{X: x, Y: y})
			}
		}
	}

	return freeCells
}

func (gs *GameSession) coordKey(x, y int32) string {
	return string(x) + "," + string(y)
}

func (gs *GameSession) getFree5x5Square() (int32, int32, Direction, bool) {
	freeSquares := gs.findAllFree5x5Squares()
	if len(freeSquares) == 0 {
		return 0, 0, Direction_UP, false
	}

	rand.Shuffle(len(freeSquares), func(i, j int) {
		freeSquares[i], freeSquares[j] = freeSquares[j], freeSquares[i]
	})

	for _, square := range freeSquares {
		centerX, centerY := square.X, square.Y

		directions := []Direction{Direction_UP, Direction_DOWN, Direction_LEFT, Direction_RIGHT}
		rand.Shuffle(len(directions), func(i, j int) {
			directions[i], directions[j] = directions[j], directions[i]
		})

		for _, dir := range directions {
			tailX, tailY := gs.getTailPosition(centerX, centerY, dir)

			if !gs.hasFoodAt(centerX, centerY) && !gs.hasFoodAt(tailX, tailY) {
				return centerX, centerY, dir, true
			}
		}
	}

	return 0, 0, Direction_UP, false
}

func (gs *GameSession) findAllFree5x5Squares() []*GameState_Coord {
	var freeSquares []*GameState_Coord

	for x := int32(0); x < gs.Config.Width; x++ {
		for y := int32(0); y < gs.Config.Height; y++ {
			if gs.is5x5SquareFree(x, y) {
				freeSquares = append(freeSquares, &GameState_Coord{X: x, Y: y})
			}
		}
	}

	return freeSquares
}

func (gs *GameSession) is5x5SquareFree(centerX, centerY int32) bool {
	for dx := int32(-2); dx <= 2; dx++ {
		for dy := int32(-2); dy <= 2; dy++ {
			x := gs.normalizeX(centerX + dx)
			y := gs.normalizeY(centerY + dy)

			if gs.isCellOccupied(x, y) {
				return false
			}
		}
	}
	return true
}

func (gs *GameSession) isCellOccupied(x, y int32) bool {
	for _, player := range gs.Players {
		if player == nil || player.Snake == nil {
			continue
		}

		absCoords := gs.getSnakeAbsoluteCoordinates(player.Snake)
		for _, coord := range absCoords {
			if coord.X == x && coord.Y == y {
				return true
			}
		}
	}
	return false
}

func (gs *GameSession) getSnakeAbsoluteCoordinates(snake *GameState_Snake) []*GameState_Coord {
	if len(snake.Points) == 0 {
		return nil
	}

	coords := make([]*GameState_Coord, 0)
	currentX, currentY := snake.Points[0].X, snake.Points[0].Y
	coords = append(coords, &GameState_Coord{X: currentX, Y: currentY})

	for i := 1; i < len(snake.Points); i++ {
		currentX = gs.normalizeX(currentX + snake.Points[i].X)
		currentY = gs.normalizeY(currentY + snake.Points[i].Y)
		coords = append(coords, &GameState_Coord{X: currentX, Y: currentY})
	}

	return coords
}

func (gs *GameSession) getTailPosition(headX, headY int32, direction Direction) (int32, int32) {
	switch direction {
	case Direction_UP:
		return gs.normalizeX(headX), gs.normalizeY(headY - 1)
	case Direction_DOWN:
		return gs.normalizeX(headX), gs.normalizeY(headY + 1)
	case Direction_LEFT:
		return gs.normalizeX(headX - 1), gs.normalizeY(headY)
	case Direction_RIGHT:
		return gs.normalizeX(headX + 1), gs.normalizeY(headY)
	default:
		return headX, headY
	}
}

func (gs *GameSession) hasFoodAt(x, y int32) bool {
	for _, food := range gs.State.Foods {
		if food.X == x && food.Y == y {
			return true
		}
	}
	return false
}

func (gs *GameSession) normalizeX(x int32) int32 {
	if x < 0 {
		return gs.Config.Width + (x % gs.Config.Width)
	}
	return x % gs.Config.Width
}

func (gs *GameSession) normalizeY(y int32) int32 {
	if y < 0 {
		return gs.Config.Height + (y % gs.Config.Height)
	}
	return y % gs.Config.Height
}

func (gs *GameSession) AddPlayer(player *GamePlayer) (*PlayerWrapper, bool) {
	switch player.Role {
	case NodeRole_NORMAL:
		headX, headY, tailDirection, found := gs.getFree5x5Square()
		if !found {
			return nil, false
		}

		snake := gs.createSnakeForPlayer(player, headX, headY, tailDirection)

		pw := &PlayerWrapper{
			Player:           player,
			Snake:            snake,
			CurrentDirection: getOppositeDirection(tailDirection),
		}

		gs.Players = append(gs.Players, pw)

		return pw, true
	case NodeRole_VIEWER:
		pw := &PlayerWrapper{
			Player:           player,
			Snake:            nil,
			CurrentDirection: Direction_RIGHT,
		}

		gs.Players = append(gs.Players, pw)

		return pw, true
	}
	return nil, false
}

func (gs *GameSession) createSnakeForPlayer(player *GamePlayer, headX, headY int32, tailDirection Direction) *GameState_Snake {
	tailX, tailY := gs.getTailPosition(headX, headY, tailDirection)

	tailOffsetX := tailX - headX
	tailOffsetY := tailY - headY

	if tailOffsetX > gs.Config.Width/2 {
		tailOffsetX -= gs.Config.Width
	} else if tailOffsetX < -gs.Config.Width/2 {
		tailOffsetX += gs.Config.Width
	}

	if tailOffsetY > gs.Config.Height/2 {
		tailOffsetY -= gs.Config.Height
	} else if tailOffsetY < -gs.Config.Height/2 {
		tailOffsetY += gs.Config.Height
	}

	points := []*GameState_Coord{
		{X: headX, Y: headY},
		{X: tailOffsetX, Y: tailOffsetY},
	}

	return &GameState_Snake{
		PlayerId:      player.Id,
		Points:        points,
		State:         GameState_Snake_ALIVE,
		HeadDirection: getOppositeDirection(tailDirection),
	}
}

func IsDirectionValid(first Direction, second Direction) bool {
	if getOppositeDirection(first) == second {
		return false
	}
	return true
}

func getOppositeDirection(dir Direction) Direction {
	switch dir {
	case Direction_UP:
		return Direction_DOWN
	case Direction_DOWN:
		return Direction_UP
	case Direction_LEFT:
		return Direction_RIGHT
	case Direction_RIGHT:
		return Direction_LEFT
	default:
		return Direction_UP
	}
}

func (gs *GameSession) isHeadColliding(head *GameState_Coord, occupiedCells map[string][]int32, currentSnakeId int32) bool {
	key := gs.coordKey(head.X, head.Y)

	if snakes, exists := occupiedCells[key]; exists {
		if len(snakes) > 1 {
			return true
		}
		if len(snakes) == 1 && snakes[0] != currentSnakeId {
			return true
		}
	}

	return false
}

func (gs *GameSession) getAllOccupiedCells() map[string][]int32 {
	occupied := make(map[string][]int32)

	for _, player := range gs.Players {
		if player == nil || player.Snake == nil || player.Snake.State != GameState_Snake_ALIVE {
			continue
		}

		snakeCoords := gs.getSnakeAbsoluteCoordinates(player.Snake)
		for _, coord := range snakeCoords {
			key := gs.coordKey(coord.X, coord.Y)
			occupied[key] = append(occupied[key], player.Snake.PlayerId)
		}
	}

	return occupied
}

func (gs *GameSession) convertSnakeToFood(snake *GameState_Snake) {
	absCoords := gs.getSnakeAbsoluteCoordinates(snake)

	for _, coord := range absCoords {
		if rand.Float32() < 0.5 {
			if !gs.hasFoodAt(coord.X, coord.Y) {
				gs.State.Foods = append(gs.State.Foods, &GameState_Coord{X: coord.X, Y: coord.Y})
			}
		}
	}
}

func (gs *GameSession) CheckCollisions() {
	occupiedCells := gs.getAllOccupiedCells()
	deadSnakes := make(map[int32]bool)

	for _, player := range gs.Players {
		if player == nil || player.Snake == nil || player.Snake.State != GameState_Snake_ALIVE {
			continue
		}

		snakeCoords := gs.getSnakeAbsoluteCoordinates(player.Snake)
		if len(snakeCoords) == 0 {
			continue
		}

		head := snakeCoords[0]
		key := gs.coordKey(head.X, head.Y)

		if gs.isHeadColliding(head, occupiedCells, player.Snake.PlayerId) {
			player.Snake.State = GameState_Snake_ZOMBIE
			deadSnakes[player.Snake.PlayerId] = true
			gs.convertSnakeToFood(player.Snake)

			gs.awardPointsForCollision(key, player.Snake.PlayerId, occupiedCells)
		}
	}

	for i := range gs.Players {
		if _, ok := deadSnakes[gs.Players[i].Player.Id]; ok {
			gs.Players[i].Snake = nil
		}
	}
}

func (gs *GameSession) awardPointsForCollision(collisionPointKey string, collidingSnakeId int32, occupiedCells map[string][]int32) {
	if snakeIDs, exists := occupiedCells[collisionPointKey]; exists {
		for _, snakeID := range snakeIDs {
			if snakeID != collidingSnakeId {
				for _, player := range gs.Players {
					if player != nil && player.Snake != nil &&
						player.Snake.State == GameState_Snake_ALIVE &&
						player.Snake.PlayerId == snakeID {
						player.Player.Score++
					}
				}
			}
		}
	}
}
