package ui

import (
	"snake-game/internal/application/game_objects"
	"snake-game/internal/domain"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type Controller struct {
	player *game_objects.Player

	currentMovement domain.Direction

	lastUpdate   time.Time
	updatePeriod time.Duration
}

func (c *Controller) GetPoints() []*domain.GameState_Coord {
	return c.player.GetPoints()
}

func (c *Controller) SetPoints(points []*domain.GameState_Coord) {
	c.player.SetPoints(points)
}

func (c *Controller) GrowPlayer() {
	points := c.player.GetPoints()
	x := *points[len(points)-1].X
	y := *points[len(points)-1].Y
	c.SetPoints(append(points, &domain.GameState_Coord{X: &x, Y: &y}))
}

func (c *Controller) SetPlayer(x, y int32) {
	c.player = game_objects.NewPlayer(x, y)
	c.currentMovement = domain.Direction_RIGHT
	c.lastUpdate = time.Now()
	c.updatePeriod = time.Millisecond * 300
}

func (c *Controller) Move() {
	c.player.Move(c.currentMovement)
}

func (c *Controller) Kill() {
	//TODO implement me
	panic("implement me")
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

	if time.Since(c.lastUpdate) >= c.updatePeriod {
		c.lastUpdate = time.Now()
		c.Move()
	}
}

func (c *Controller) DrawPlayer(screen *ebiten.Image, grid *domain.Grid) {
	c.player.Draw(screen, grid)
}
