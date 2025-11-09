package ui

import (
	"snake-game/internal/domain"

	"github.com/hajimehoshi/ebiten/v2"
)

type Controller struct {
	player          *domain.PlayerWrapper
	currentMovement domain.Direction
}

func (c *Controller) Player() *domain.GamePlayer {
	return c.player.Player
}

func (c *Controller) Snake() *domain.GameState_Snake {
	return c.player.Snake
}

func (c *Controller) SetIpAndPort(ip string, port int32) {
	c.player.Player.IpAddress = ip
	c.player.Player.Port = port
}

func (c *Controller) Id() int32 {
	return c.player.Player.Id
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
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyW):
		c.currentMovement = domain.Direction_UP
	case ebiten.IsKeyPressed(ebiten.KeyA):
		c.currentMovement = domain.Direction_LEFT
	case ebiten.IsKeyPressed(ebiten.KeyD):
		c.currentMovement = domain.Direction_RIGHT
	case ebiten.IsKeyPressed(ebiten.KeyS):
		c.currentMovement = domain.Direction_DOWN
	}
}
