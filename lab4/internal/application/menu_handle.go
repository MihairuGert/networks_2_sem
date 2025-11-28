package application

import (
	"snake-game/internal/application/network"
	"snake-game/internal/application/ui"
	"snake-game/internal/domain"
	"time"
)

func (g *Game) handleNewGame() {
	g.startGame()
	g.startNetwork()

	g.state = Play
}

func (g *Game) startGame() {
	config, err := ui.ParseConfig("conf.yaml")
	if err != nil {
		panic(err)
	}
	g.GameSession = domain.NewGameSession(config, float32(screenWidthGlobal), float32(screenHeightGlobal))
	g.setUpRenderer()

	g.GameSession.BecomeMaster()

	controller := ui.Controller{}
	g.GameSession.SetMyID(g.GameSession.GetFreePlayerId())
	gp := domain.GamePlayer{
		Name:      "me",
		Id:        0,
		IpAddress: "",
		Port:      0,
		Role:      0,
		Type:      0,
		Score:     0,
	}
	player := g.addPlayer(&gp)
	controller.SetPlayer(player)
	g.addController(controller)

	// todo should be taken from config
	g.lastFoodSpawnTime = time.Now()
	g.foodSpawnInt = time.Second * 3
}

func (g *Game) setUpRenderer() {
	renderer := ui.GameSessionRenderer{ScreenWidth: float32(screenWidthGlobal), ScreenHeight: float32(screenHeightGlobal)}
	g.Renderer = &renderer
	g.Renderer.SetGridImage(g.GameSession.Grid)
}

func (g *Game) startNetwork() {
	g.networkManager = network.NewNetworkManager()
	g.goroutinePool.Go(g.startAnnouncement)
	g.startListening()
}

func (g *Game) handleConnect() {
	g.availableGames = make(map[string]AvailableGame)
	g.startNetwork()

	var game AvailableGame
	for {
		err := g.handleIncomingMessages()
		if err != nil {
			continue
		}
		err = g.discoverGame()
		if err != nil {
			continue
		}
		game, err = g.findGame()
		if err != nil {
			continue
		}
		break
	}
	g.networkManager.StartAckDaemonWithDuration(time.Duration(game.Msg.Config.StateDelayMs/10) * time.Millisecond)
	seqNum := g.JoinGame(game.Addr(), game.Msg.GetGameName(), game.Msg.GetCanJoin())
	var ok bool
	var ackMsg *domain.GameMessage
	for {
		ok, ackMsg = g.networkManager.CheckAck(seqNum)
		if ok {
			break
		}
		err := g.handleIncomingMessages()
		if err != nil {
			continue
		}
		time.Sleep(time.Duration(game.Msg.Config.StateDelayMs/100) * time.Millisecond)
	}
	g.GameSession = domain.NewGameSession(game.Msg.Config, float32(screenWidthGlobal), float32(screenHeightGlobal))
	g.GameSession.SetMyID(ackMsg.ReceiverId)
	g.setUpRenderer()
	g.GameSession.BecomeNormal()

	controller := ui.Controller{}
	g.myPlayer = &domain.PlayerWrapper{}
	g.myPlayer.CurrentDirection = domain.Direction_RIGHT
	controller.SetPlayer(g.myPlayer)
	g.addController(controller)

	g.state = Play
}

func (g *Game) handleExit() {
	g.state = End
	g.shutdownTime = time.Now()
	g.lastFlickTime = time.Now()
	g.flickerInt = 25 * time.Millisecond
	g.finalMsg = ui.NewText("", 24, 100, 100)
}
