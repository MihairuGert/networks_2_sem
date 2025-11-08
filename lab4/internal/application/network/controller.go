package network

import (
	"snake-game/internal/application/game_objects"
	"snake-game/internal/domain"

	"github.com/hajimehoshi/ebiten/v2"
)

type Controller struct {
	player *game_objects.Player
	domain.GamePlayer

	currentMovement domain.Direction
}

func (c *Controller) SetId(id int32) {
	c.Id = id
}

func (c *Controller) GrowPlayer() {
	c.player.Grow()
}

func (c *Controller) GetPoints() []*domain.GameState_Coord {
	return c.player.GetPoints()
}

func (c *Controller) SetPoints(points []*domain.GameState_Coord) {
	c.player.SetPoints(points)
}

func (c *Controller) SetPlayer(x, y int32) {
	c.player = game_objects.NewPlayer(x, y)
	c.currentMovement = domain.Direction_RIGHT
}

func (c *Controller) Move() {
	c.player.Move(c.currentMovement)
}

func (c *Controller) Kill() {
	//TODO implement me
	panic("implement me")
}

func (c *Controller) Update() {

}

func (c *Controller) DrawPlayer(screen *ebiten.Image, grid *domain.Grid) {
	c.player.Draw(screen, grid)
}
