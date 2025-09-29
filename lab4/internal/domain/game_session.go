package domain

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
