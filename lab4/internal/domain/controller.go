package domain

import "github.com/hajimehoshi/ebiten/v2"

type Controller interface {
	GetPoints() []*GameState_Coord
	SetPoints(points []*GameState_Coord)
	SetPlayer(x, y int32, name string, id int32)
	Move()
	Kill()
	Update()
	DrawPlayer(screen *ebiten.Image, grid *Grid)
	GrowPlayer()
	SetId(id int32)
	Id() int32
}
