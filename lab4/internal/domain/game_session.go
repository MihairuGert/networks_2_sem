package domain

type GameSession struct {
	Grid *Grid
	GameConfig
	GamePlayers
	GameState
}
