package domain

type PlayerWrapper struct {
	Player *GamePlayer
	Snake  *GameState_Snake

	CurrentDirection Direction
}

func NewPlayer(x, y int32, gp *GamePlayer) *PlayerWrapper {
	player := gp

	points := []*GameState_Coord{
		{X: x, Y: y},
		{X: -1, Y: 0},
	}

	snake := &GameState_Snake{
		PlayerId:      gp.Id,
		Points:        points,
		State:         GameState_Snake_ALIVE,
		HeadDirection: Direction_RIGHT,
	}

	return &PlayerWrapper{
		Player: player,
		Snake:  snake,
	}
}

func (pw *PlayerWrapper) GetPoints() []*GameState_Coord {
	return pw.Snake.Points
}

func (pw *PlayerWrapper) SetPoints(points []*GameState_Coord) {
	pw.Snake.Points = points
}

func (pw *PlayerWrapper) Move() {
	if pw.Snake == nil {
		return
	}
	for i := len(pw.Snake.Points) - 1; i > 1; i-- {
		pw.Snake.Points[i].X = pw.Snake.Points[i-1].X
		pw.Snake.Points[i].Y = pw.Snake.Points[i-1].Y
	}

	pw.Snake.HeadDirection = pw.CurrentDirection
	switch pw.Snake.HeadDirection {
	case Direction_UP:
		pw.Snake.Points[0].Y -= 1
		pw.Snake.Points[1].Y = 1
		pw.Snake.Points[1].X = 0
	case Direction_DOWN:
		pw.Snake.Points[0].Y += 1
		pw.Snake.Points[1].Y = -1
		pw.Snake.Points[1].X = 0
	case Direction_RIGHT:
		pw.Snake.Points[0].X += 1
		pw.Snake.Points[1].X = -1
		pw.Snake.Points[1].Y = 0
	case Direction_LEFT:
		pw.Snake.Points[0].X -= 1
		pw.Snake.Points[1].X = 1
		pw.Snake.Points[1].Y = 0
	}
}

func (pw *PlayerWrapper) Grow() {
	points := pw.GetPoints()
	x := points[len(points)-1].X
	y := points[len(points)-1].Y
	pw.SetPoints(append(points, &GameState_Coord{X: x, Y: y}))
	pw.Player.Score++
}
