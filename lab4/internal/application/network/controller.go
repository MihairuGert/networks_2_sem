package network

import (
	"snake-game/internal/domain"

	"github.com/hajimehoshi/ebiten/v2"
)

type Controller struct {
	player          *domain.PlayerWrapper
	currentMovement domain.Direction
}

func (c *Controller) Id() int32 {
	return c.player.Player.Id
}

func (c *Controller) SetIpAndPort(ip string, port int32) {
	c.player.Player.IpAddress = ip
	c.player.Player.Port = port
}

func (c *Controller) SetId(id int32) {
	c.player.Player.Id = id
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

func (c *Controller) SetPlayer(x, y int32, name string, id int32) {
	c.player = domain.NewPlayer(x, y, name, id)
	c.currentMovement = domain.Direction_RIGHT
}

func (c *Controller) Move() {
	c.player.Move(c.currentMovement)
}

func (c *Controller) Kill() {
	c.player.Snake.State = domain.GameState_Snake_ZOMBIE
}

func (c *Controller) Update() {

}

func (c *Controller) DrawPlayer(screen *ebiten.Image, grid *domain.Grid) {
	c.player.Draw(screen, grid)
}
