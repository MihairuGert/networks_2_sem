package application

import (
	"context"
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
	Play
	GameSelection
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
	controller  ui.Controller
	myPlayer    *domain.PlayerWrapper

	networkManager *network.Manager
	ticker         *time.Ticker
	goroutinePool  *errgroup.Group
	shouldStop     bool

	state               gameState
	availableGames      map[string]AvailableGame
	availableGamesMutex sync.Mutex

	// Used in fancy exit window.
	shutdownTime  time.Time
	finalMsg      *ui.Text
	flickerInt    time.Duration
	lastFlickTime time.Time
}

func (g *Game) Init() error {
	err := g.setUpWindow()
	if err != nil {
		return err
	}

	g.goroutinePool, _ = errgroup.WithContext(context.Background())
	g.setupMenu()
	g.state = Menu
	return nil
}

func (g *Game) setUpWindow() error {
	ebiten.SetWindowSize(int(screenWidthGlobal), int(screenHeightGlobal))
	ebiten.SetWindowTitle("Mihairu's Snake Game")
	_, icon, err := ebitenutil.NewImageFromFile(texturesPath + "app_icon.jpeg")
	if err != nil {
		return err
	}
	icons := []image.Image{icon}
	ebiten.SetWindowIcon(icons)
	return nil
}

func (g *Game) addController(controller ui.Controller) {
	g.controller = controller
}

//func (g *Game) addPlayer(gp *domain.GamePlayer) *domain.PlayerWrapper {
//	player := domain.NewPlayer(0, 0, gp)
//	player.CurrentDirection = domain.Direction_RIGHT
//	g.GameSession.Players = append(g.GameSession.Players, player)
//	return player
//}

func (g *Game) Start() error {
	if err := ebiten.RunGame(g); err != nil {
		return err
	}
	return nil
}

func (g *Game) setState() {
	g.GameSession.State.StateOrder = int32(g.GameSession.CurrentStateNum())
	var players []*domain.GamePlayer
	var snakes []*domain.GameState_Snake
	for _, controller := range g.GameSession.Players {
		if controller.Player != nil {
			players = append(players, controller.Player)
		}
		if controller.Snake != nil {
			snakes = append(snakes, controller.Snake)
		}
	}
	g.GameSession.State.Snakes = snakes
	g.GameSession.State.Players = &domain.GamePlayers{Players: players}
}
