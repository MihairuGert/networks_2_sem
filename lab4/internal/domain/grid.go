package domain

type Grid struct {
	Width      float32
	Height     float32
	RectHeight float32
	RectWidth  float32
}

func NewGrid(width, height float32) *Grid {
	return &Grid{Width: width, Height: height}
}

func (g *Grid) getSquareSize(screenWidth, screenHeight float32) {
	g.RectWidth = screenWidth / g.Width
	g.RectHeight = screenHeight / g.Height
}
