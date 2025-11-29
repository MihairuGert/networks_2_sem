package application

import (
	"fmt"
	"snake-game/internal/application/network"
	"snake-game/internal/application/ui"
	"snake-game/internal/domain"
	"strings"
	"time"
)

func (g *Game) handleNewGame() {
	g.startGame()
	g.startNetwork()

	g.state = Play
}

func (g *Game) handleExitGame() {
	g.endGame()
	g.stopNetwork()

	g.state = Menu
}

func (g *Game) endGame() {

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
		Type:      domain.PlayerType_HUMAN,
		Score:     0,
	}
	player, _ := g.GameSession.AddPlayer(&gp)
	controller.SetPlayer(player)
	g.addController(controller)
	player.Player.Role = domain.NodeRole_MASTER

	// todo should be taken from config
	g.lastFoodSpawnTime = time.Now()
	g.foodSpawnInt = time.Second * 3
}

func (g *Game) setUpRenderer() {
	renderer := ui.GameSessionRenderer{
		ScreenWidth:  float32(screenWidthGlobal),
		ScreenHeight: float32(screenHeightGlobal),
		PlayerList:   ui.NewPlayerList(float64(screenWidthGlobal-300), 50, 25),
		ExitButton:   ui.NewTextButton("Exit", 10, 10, screenHeightGlobal-200, 50, 50, g.handleExitGame),
	}
	g.Renderer = &renderer
	g.Renderer.SetGridImage(g.GameSession.Grid)
}

func (g *Game) startNetwork() {
	g.shouldStop = false
	g.networkManager = network.NewNetworkManager(&g.shouldStop)
	g.goroutinePool.Go(g.startAnnouncement)
	g.startListening()
}

func (g *Game) stopNetwork() {
	g.shouldStop = true
}

func (g *Game) handleConnect() {
	g.availableGames = make(map[string]AvailableGame)
	g.startNetwork()

	var game AvailableGame
	for {
		err := g.handleIncomingMessages()
		if err != nil {
			println(err)
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
	viewOnly := true
	if game.Msg.GetCanJoin() == true {
		fmt.Println("Do you want to play the game?[Y/N]")
		char := ""
		fmt.Scan(&char)
		switch strings.ToLower(char) {
		case "y":
			viewOnly = false
		default:
			viewOnly = true
		}
	}
	// todo move to init?
	g.networkManager.StartAckDaemonWithDuration(time.Duration(game.Msg.Config.StateDelayMs/10) * time.Millisecond)
	seqNum := g.JoinGame(game.Addr(), game.Msg.GetGameName(), viewOnly)
	var ok bool
	var ackMsg *domain.GameMessage
	for {
		ok, ackMsg = g.networkManager.CheckAck(seqNum)
		if ok {
			break
		}
		err := g.handleIncomingMessages()
		if err != nil {
			fmt.Printf("Got error:%s\n", err.Error())
			return
		}
		time.Sleep(time.Duration(game.Msg.Config.StateDelayMs/100) * time.Millisecond)
	}
	g.GameSession = domain.NewGameSession(game.Msg.Config, float32(screenWidthGlobal), float32(screenHeightGlobal))
	g.GameSession.SetMyID(ackMsg.ReceiverId)
	g.setUpRenderer()
	g.GameSession.Node.SetMasterAddr(game.Addr())
	if viewOnly {
		g.GameSession.BecomeViewer()
	} else {
		g.GameSession.BecomeNormal()
	}

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
