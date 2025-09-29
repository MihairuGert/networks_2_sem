package ui

import (
	"snake-game/internal/application/game_objects"
	"snake-game/internal/domain"

	"github.com/hajimehoshi/ebiten/v2"
)

type Controller struct {
	player *game_objects.Player
}

func (c *Controller) Move(currentMovement []domain.Direction) {
	c.player.Move(currentMovement)
}

func (c *Controller) Kill() {
	//TODO implement me
	panic("implement me")
}

func (c *Controller) Update() {
	var currentMovement []domain.Direction
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyW):
		currentMovement = append(currentMovement, domain.Direction_UP)
	case ebiten.IsKeyPressed(ebiten.KeyA):
		currentMovement = append(currentMovement, domain.Direction_LEFT)
	case ebiten.IsKeyPressed(ebiten.KeyD):
		currentMovement = append(currentMovement, domain.Direction_RIGHT)
	case ebiten.IsKeyPressed(ebiten.KeyS):
		currentMovement = append(currentMovement, domain.Direction_DOWN)
	}
	c.Move(currentMovement)
}
