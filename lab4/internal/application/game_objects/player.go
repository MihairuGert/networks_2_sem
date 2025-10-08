package game_objects

import (
	"image/color"
	"snake-game/internal/domain"

	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	snake    domain.GameState_Snake
	velocity int32 // cells/sec.
}

func (player *Player) GetPoints() []*domain.GameState_Coord {
	return player.snake.Points
}

func (player *Player) SetPoints(points []*domain.GameState_Coord) {
	player.snake.Points = points
}

func NewPlayer(x, y int32) *Player {
	player := &Player{}
	player.velocity = 1

	points := make([]*domain.GameState_Coord, 0)
	lcoords := make([]int32, 8)
	lcoords[0] = x
	lcoords[1] = y
	lcoords[2] = -1
	lcoords[3] = 0

	lcoords[4] = -1
	lcoords[5] = 0

	lcoords[6] = -1
	lcoords[7] = 0

	points = append(points, &domain.GameState_Coord{X: &lcoords[0], Y: &lcoords[1]})
	points = append(points, &domain.GameState_Coord{X: &lcoords[2], Y: &lcoords[3]})
	points = append(points, &domain.GameState_Coord{X: &lcoords[4], Y: &lcoords[5]})
	points = append(points, &domain.GameState_Coord{X: &lcoords[6], Y: &lcoords[7]})

	player.snake = domain.GameState_Snake{Points: points}
	curDir := domain.Direction_RIGHT
	player.snake.HeadDirection = &curDir

	return player
}

func (player *Player) Move(direction domain.Direction) {
	for i := len(player.snake.Points) - 1; i > 1; i-- {
		*player.snake.Points[i].X = *player.snake.Points[i-1].X
		*player.snake.Points[i].Y = *player.snake.Points[i-1].Y
	}

	*player.snake.HeadDirection = direction
	switch *player.snake.HeadDirection {
	case domain.Direction_UP:
		*player.snake.Points[0].Y -= player.velocity
		*player.snake.Points[1].Y = 1
		*player.snake.Points[1].X = 0
	case domain.Direction_DOWN:
		*player.snake.Points[0].Y += player.velocity
		*player.snake.Points[1].Y = -1
		*player.snake.Points[1].X = 0
	case domain.Direction_RIGHT:
		*player.snake.Points[0].X += player.velocity
		*player.snake.Points[1].X = -1
		*player.snake.Points[1].Y = 0
	case domain.Direction_LEFT:
		*player.snake.Points[0].X -= player.velocity
		*player.snake.Points[1].X = 1
		*player.snake.Points[1].Y = 0
	default:

	}
}

func (player *Player) Draw(screen *ebiten.Image, grid *domain.Grid) {
	rectImage := ebiten.NewImage(int(grid.RectWidth), int(grid.RectHeight))
	rectImage.Fill(color.RGBA{R: 255, G: 0, B: 0, A: 255})

	curX := float64(*player.snake.Points[0].X) * float64(grid.RectWidth)
	curY := float64(*player.snake.Points[0].Y) * float64(grid.RectHeight)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(curX, curY)
	screen.DrawImage(rectImage, opts)

	for i := 1; i < len(player.snake.Points); i++ {
		curX = curX + float64(grid.RectWidth)*float64(*player.snake.Points[i].X)
		curY = curY + float64(grid.RectHeight)*float64(*player.snake.Points[i].Y)

		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(curX, curY)
		screen.DrawImage(rectImage, opts)
	}
}
