package domain

type Grid struct {
	Width      float32
	Height     float32
	RectHeight float32
	RectWidth  float32
}

func NewGrid(width, height, screenWidth, screenHeight float32) *Grid {
	grid := Grid{Width: width, Height: height}
	grid.getSquareSize(screenWidth, screenHeight)
	return &grid
}

func (g *Grid) getSquareSize(screenWidth, screenHeight float32) {
	g.RectWidth = screenWidth / g.Width
	g.RectHeight = screenHeight / g.Height
}
