package domain

type Grid struct {
	cells [][]int

	Width      int
	Height     int
	RectHeight float32
	RectWidth  float32
}

const Scale = 0.5

func NewGrid(width, height int, screenWidth, screenHeight float32) *Grid {
	grid := Grid{Width: width, Height: height, cells: make([][]int, height)}
	for i := 0; i < height; i++ {
		grid.cells[i] = make([]int, width)
	}
	grid.getSquareSize(screenWidth, screenHeight)
	return &grid
}

func (g *Grid) getSquareSize(screenWidth, screenHeight float32) {
	g.RectWidth = (screenWidth / float32(g.Width)) * Scale
	g.RectHeight = g.RectWidth

	// leaving this for future if decide to make grid rectangle-like.
	// screenHeight / g.Height
}
