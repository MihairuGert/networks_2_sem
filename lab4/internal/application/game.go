package application

import (
	"image"
	_ "image/jpeg"
	"snake-game/internal/application/network"
	"snake-game/internal/application/ui"
	"snake-game/internal/domain"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/sync/errgroup"
)

var (
	screenWidthGlobal  = 960
	screenHeightGlobal = 720
)

const texturesPath = "./textures/"

type gameState int

const (
	Menu gameState = iota
	Connect
	Play
	End
)

type AvailableGame struct {
	Msg  *domain.GameAnnouncement
	addr string
}

func (a *AvailableGame) Addr() string {
	return a.addr
}

type Game struct {
	Renderer *ui.GameSessionRenderer
	Menu     *ui.Menu

	GameSession *domain.GameSession
	controllers map[int]domain.Controller

	handleChannel  chan network.Msg
	networkManager *network.Manager
	goroutinePool  *errgroup.Group

	state               gameState
	availableGames      map[string]AvailableGame
	availableGamesMutex sync.Mutex

	// Used in fancy exit window.
	shutdownTime  time.Time
	finalMsg      *ui.Text
	flickerInt    time.Duration
	lastFlickTime time.Time

	// todo move it to game config!
	foodSpawnInt      time.Duration
	lastFoodSpawnTime time.Time
}

func (g *Game) Init() error {
	err := g.setUpWindow()
	if err != nil {
		return err
	}

	g.setupMenu()
	g.state = Menu
	return nil
}

func (g *Game) setUpWindow() error {
	ebiten.SetWindowSize(screenWidthGlobal, screenHeightGlobal)
	ebiten.SetWindowTitle("Mihairu's Snake Game")
	_, icon, err := ebitenutil.NewImageFromFile(texturesPath + "app_icon.jpeg")
	if err != nil {
		return err
	}
	icons := []image.Image{icon}
	ebiten.SetWindowIcon(icons)
	return nil
}

func (g *Game) addPlayer(c domain.Controller) int {
	id := g.GameSession.GetFreePlayerId()
	c.SetId(int32(id))
	g.controllers[id] = c
	return id
}

func (g *Game) Start() error {
	if err := ebiten.RunGame(g); err != nil {
		return err
	}
	return nil
}
