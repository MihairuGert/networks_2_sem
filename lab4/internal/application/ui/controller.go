package ui

import (
	"snake-game/internal/domain"

	"github.com/hajimehoshi/ebiten/v2"
)

type Controller struct {
	player *domain.PlayerWrapper
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

func (c *Controller) SetPlayer(player *domain.PlayerWrapper) {
	c.player = player
}

func (c *Controller) Move() {
	c.player.Move()
}

func (c *Controller) Direction() domain.Direction {
	return c.player.CurrentDirection
}

func (c *Controller) Kill() {
	c.player.Snake.State = domain.GameState_Snake_ZOMBIE
}

func (c *Controller) Update() {
	if c.player == nil {
		return
	}
	chosenDir := domain.Direction_RIGHT

	switch {
	case ebiten.IsKeyPressed(ebiten.KeyW):
		chosenDir = domain.Direction_UP
	case ebiten.IsKeyPressed(ebiten.KeyA):
		chosenDir = domain.Direction_LEFT
	case ebiten.IsKeyPressed(ebiten.KeyD):
		chosenDir = domain.Direction_RIGHT
	case ebiten.IsKeyPressed(ebiten.KeyS):
		chosenDir = domain.Direction_DOWN
	default:
		return
	}

	if domain.IsDirectionValid(c.player.CurrentDirection, chosenDir) {
		c.player.CurrentDirection = chosenDir
	}
}
