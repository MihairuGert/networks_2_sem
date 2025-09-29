package game_objects

import "snake-game/internal/domain"

type Player struct {
	snake    snake
	velocity float32
}

func (player *Player) Move(direction []domain.Direction) {
	player.snake.HeadDirection = &direction[0]
}
