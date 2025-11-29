package application

import (
	"errors"
	"fmt"
	"snake-game/internal/application/network"
	"snake-game/internal/domain"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
)

func (g *Game) startAnnouncement() error {
	for {
		select {
		case <-g.ticker.C:
			if g.shouldStop {
				return nil
			}
			if g.GameSession == nil || g.GameSession.Node.Role() != domain.NodeRole_MASTER {
				continue
			}
			err2 := g.sendAnnouncementTo(network.MulticastAddress)
			if err2 != nil {
				return err2
			}
		default:
			if g.shouldStop {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (g *Game) sendAnnouncementTo(addr string) error {
	announcementMsg := &domain.GameMessage{
		MsgSeq:     g.networkManager.MsgSeq(),
		SenderId:   -1,
		ReceiverId: -1,
		Type: &domain.GameMessage_Announcement{
			Announcement: &domain.GameMessage_AnnouncementMsg{
				Games: []*domain.GameAnnouncement{{
					Players:  g.GameSession.State.Players,
					Config:   g.GameSession.Config,
					CanJoin:  true,
					GameName: "asd",
				}},
			},
		},
	}

	data, err := proto.Marshal(announcementMsg)
	if err != nil {
		return err
	}
	err = g.networkManager.SendMsg(&data, addr)
	if err != nil {
		return err
	}
	return nil
}

func (g *Game) discoverGame() error {
	discoverMsg := &domain.GameMessage{
		MsgSeq:     g.networkManager.MsgSeq(),
		SenderId:   -1,
		ReceiverId: -1,
		Type: &domain.GameMessage_Discover{
			Discover: &domain.GameMessage_DiscoverMsg{},
		},
	}

	data, err := proto.Marshal(discoverMsg)
	if err != nil {
		return err
	}

	err = g.networkManager.SendMsg(&data, network.MulticastAddress)
	if err != nil {
		return err
	}
	return nil
}

func (g *Game) JoinGame(masterAddr string, gameName string, viewOnly bool) int64 {
	res := g.networkManager.MsgSeq()
	role := domain.NodeRole_NORMAL
	if viewOnly {
		role = domain.NodeRole_VIEWER
	}
	joinMsg := &domain.GameMessage{
		MsgSeq:     res,
		SenderId:   -1,
		ReceiverId: -1,
		Type: &domain.GameMessage_Join{
			Join: &domain.GameMessage_JoinMsg{
				PlayerType:    0,
				GameName:      gameName,
				PlayerName:    "The Swan",
				RequestedRole: role,
			},
		},
	}

	data, err := proto.Marshal(joinMsg)
	if err != nil {
		fmt.Printf("Failed to marshal join: %v\n", err)
		return -1
	}

	err = g.networkManager.SendMsg(&data, masterAddr)
	if err != nil {
		fmt.Printf("Failed to send join: %v\n", err)
		return -1
	}
	addr, err := network.StringToAddr(masterAddr)
	if err != nil {
		return 0
	}
	g.networkManager.NeedAck(network.NewMsg(data, addr), res, false)
	return res
}

func (g *Game) startListening() {
	g.goroutinePool.Go(g.networkManager.ListenMulticast)
	g.goroutinePool.Go(g.networkManager.ListenUnicast)
}

func (g *Game) handleIncomingMessages() error {
	if g.GameSession != nil {
		err := g.checkPlayersConnection()
		if err != nil {
			return err
		}
	}
	messages := g.networkManager.GetUnreadMessages()
	for _, msg := range messages {
		var gameMsg domain.GameMessage
		if err := proto.Unmarshal(msg.Data(), &gameMsg); err != nil {
			continue
		}
		switch gameMsg.Type.(type) {
		case *domain.GameMessage_Discover:
			err := g.handleDiscover(&gameMsg, msg.Addr().String())
			if err != nil {
				return err
			}
		case *domain.GameMessage_Announcement:
			err := g.handleAnnouncement(&gameMsg, msg.Addr().String())
			if err != nil {
				return err
			}
		case *domain.GameMessage_Join:
			err := g.handleJoin(&gameMsg, msg.Addr().String())
			if err != nil {
				return err
			}
		case *domain.GameMessage_Ack:
			g.networkManager.SetAck(gameMsg.MsgSeq, &gameMsg)
		case *domain.GameMessage_State:
			if g.GameSession == nil {
				continue
			}
			if g.GameSession.Node.Role() == domain.NodeRole_MASTER {
				continue
			}
			state := gameMsg.GetState().State
			g.Renderer.Update(state.Players.Players)
			g.GameSession.State = state
		case *domain.GameMessage_Steer:
			if g.GameSession == nil {
				continue
			}
			if g.GameSession.Node.Role() != domain.NodeRole_MASTER {
				continue
			}
			err := g.handleSteer(&gameMsg)
			if err != nil {
				continue
			}
		case *domain.GameMessage_Error:
			g.networkManager.SetErr(gameMsg.MsgSeq, &gameMsg)
			return errors.New(gameMsg.GetError().ErrorMessage)
		case *domain.GameMessage_Ping:
			return nil
		case *domain.GameMessage_RoleChange:
			err := g.handleRoleChg(&gameMsg, msg.Addr().String())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Game) handleRoleChg(msg *domain.GameMessage, srcAddr string) error {
	if g.GameSession == nil {
		return nil
	}
	rchg := msg.GetRoleChange()
	switch {
	case rchg.ReceiverRole == domain.NodeRole_DEPUTY:
		g.GameSession.BecomeDeputy()
	}
	err := g.sendAckTo(msg, srcAddr)
	if err != nil {
		return err
	}
	return nil
}

func (g *Game) handleJoin(msg *domain.GameMessage, srcAddr string) error {
	if g.state != Play && g.GameSession.Node.Role() != domain.NodeRole_MASTER {
		return nil
	}

	id := g.GameSession.GetFreePlayerId()
	ipAddress, port := GetIpAndPort(srcAddr)

	gp := domain.GamePlayer{
		Name:      msg.GetJoin().GetPlayerName(),
		Id:        id,
		IpAddress: ipAddress,
		Port:      port,
		Role:      msg.GetJoin().RequestedRole,
		Type:      domain.PlayerType_HUMAN,
		Score:     0,
	}

	_, canJoin := g.GameSession.AddPlayer(&gp)

	if canJoin {
		msg.ReceiverId = id
		err := g.sendAckTo(msg, srcAddr)
		if err != nil {
			return err
		}
		err = g.ChooseDeputy()
		if err != nil {
			return err
		}
		return nil
	}

	err := g.sendErrorTo(msg, srcAddr)
	if err != nil {
		return err
	}
	return nil
}

func (g *Game) checkPlayersToPing() {
	needToPing := g.networkManager.GetWhoSentLessThan(time.Duration(g.GameSession.Config.StateDelayMs/10) * time.Millisecond)
	if len(needToPing) == 0 {
		return
	}
	for _, playerAddr := range needToPing {
		g.sendPing(playerAddr)
	}
}

func (g *Game) checkPlayersConnection() error {
	if g.GameSession == nil || g.GameSession.Config == nil || g.GameSession.State == nil {
		return nil
	}
	disconnectedPlayers := g.networkManager.GetWhoRecvLessThan(time.Duration(float64(g.GameSession.Config.StateDelayMs)*0.8) * time.Millisecond)
	if len(disconnectedPlayers) == 0 {
		return nil
	}
	for _, playerAddr := range disconnectedPlayers {
		switch g.GameSession.Node.Role() {
		case domain.NodeRole_NORMAL:
			deputy := g.getDeputy()
			if deputy == nil {
				g.handleExitGame()
				return Exit
			}
			deputyAddress := formatAddress(deputy.IpAddress, deputy.Port)
			g.GameSession.Node.SetMasterAddr(deputyAddress)
		case domain.NodeRole_MASTER:
			deputy := g.getDeputy()
			g.removePlayer(playerAddr)
			if deputy == nil {
				g.ChooseDeputy()
				return nil
			}
			deputyAddress := formatAddress(deputy.IpAddress, deputy.Port)
			if deputyAddress == playerAddr {
				g.ChooseDeputy()
			}
		case domain.NodeRole_DEPUTY:
			if playerAddr == g.GameSession.Node.MasterAddr() {
				g.GameSession.BecomeMaster()
				g.reformWrappers()
				g.removePlayer(playerAddr)
				err := g.ChooseDeputy()
				if err != nil {
					return err
				}
			}
		case domain.NodeRole_VIEWER:

		}
	}
	return nil
}

func (g *Game) removePlayer(playerAddr string) {
	for i := range g.GameSession.Players {
		if formatAddress(g.GameSession.Players[i].Player.IpAddress, g.GameSession.Players[i].Player.Port) == playerAddr {
			if g.GameSession.Players[i].Player.Role == domain.NodeRole_VIEWER {
				continue
			}
			g.GameSession.Players = append(g.GameSession.Players[:i], g.GameSession.Players[i+1:]...)
			break
		}
	}
}

func (g *Game) ChooseDeputy() error {
	ind := g.GameSession.ChooseDeputy()
	if ind >= 0 {
		dip := g.GameSession.Players[ind].Player.IpAddress
		dPort := g.GameSession.Players[ind].Player.Port
		err := g.sendRoleChangeMsg(formatAddress(dip, dPort))
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Game) getDeputy() *domain.GamePlayer {
	if g.GameSession == nil || g.GameSession.State == nil || g.GameSession.State.Players == nil {
		return nil
	}
	for _, player := range g.GameSession.State.Players.Players {
		if player.Role == domain.NodeRole_DEPUTY {
			return player
		}
	}
	return nil
}

func (g *Game) reformWrappers() {
	var players []*domain.PlayerWrapper
	for _, player := range g.GameSession.State.Players.Players {
		players = append(players, &domain.PlayerWrapper{Player: player})
	}
	for _, snake := range g.GameSession.State.Snakes {
		for i := range players {
			id := players[i].Player.Id
			if id == g.GameSession.MyID() {
				g.controller.SetPlayer(players[i])
			}
			if snake.PlayerId == id {
				players[i].Snake = snake
				players[i].CurrentDirection = snake.GetHeadDirection()
				break
			}
		}
	}
	g.myPlayer = nil
	g.GameSession.Players = players
}

func (g *Game) handleSteer(msg *domain.GameMessage) error {
	steer := msg.GetSteer()
	if steer == nil {
		return errors.New("invalid steer")
	}

	for i, _ := range g.GameSession.Players {
		if g.GameSession.Players[i].Player.Id != msg.SenderId {
			continue
		}
		if domain.IsDirectionValid(g.GameSession.Players[i].CurrentDirection, steer.GetDirection()) {
			g.GameSession.Players[i].CurrentDirection = steer.GetDirection()
		}
	}
	return nil
}

func (g *Game) handleDiscover(msg *domain.GameMessage, addr string) error {
	if g.GameSession == nil {
		return nil
	}
	if g.GameSession.Node.Role() != domain.NodeRole_MASTER {
		return nil
	}
	//fmt.Printf("Received discover from %s, responding with announcement\n", addr)
	err := g.sendAnnouncementTo(addr)
	if err != nil {
		return err
	}
	return nil
}

func (g *Game) handleAnnouncement(msg *domain.GameMessage, srcAddr string) error {
	if g.state != Menu {
		return nil
	}

	announcement := msg.GetAnnouncement()
	if announcement == nil {
		return errors.New("invalid announcement")
	}

	//fmt.Printf("Received announcement from %s with %d games\n",
	//srcAddr, len(announcement.Games))

	g.availableGamesMutex.Lock()
	g.availableGames[srcAddr] = AvailableGame{Msg: announcement.Games[0], addr: srcAddr}
	g.availableGamesMutex.Unlock()

	return nil
}

func (g *Game) sendAckTo(originalMsg *domain.GameMessage, dest string) error {
	ackMsg := &domain.GameMessage{
		MsgSeq:     originalMsg.MsgSeq,
		SenderId:   originalMsg.GetSenderId(),
		ReceiverId: originalMsg.GetReceiverId(),
		Type: &domain.GameMessage_Ack{
			Ack: &domain.GameMessage_AckMsg{},
		},
	}

	data, err := proto.Marshal(ackMsg)
	if err != nil {
		return err
	}
	err = g.networkManager.SendMsg(&data, dest)
	if err != nil {
		return err
	}
	return nil
}

func (g *Game) sendErrorTo(msg *domain.GameMessage, dest string) error {
	errMsg := &domain.GameMessage{
		MsgSeq:     msg.MsgSeq,
		SenderId:   g.GameSession.MyID(),
		ReceiverId: -1,
		Type: &domain.GameMessage_Error{
			Error: &domain.GameMessage_ErrorMsg{ErrorMessage: "Not enough space on the grid. Try to become a viewer then."},
		},
	}

	data, err := proto.Marshal(errMsg)
	if err != nil {
		return err
	}
	err = g.networkManager.SendMsg(&data, dest)
	if err != nil {
		return err
	}
	addr, err := network.StringToAddr(dest)
	if err != nil {
		return err
	}
	g.networkManager.NeedAck(network.NewMsg(data, addr), msg.MsgSeq, true)
	return nil
}

func (g *Game) sendState() error {
	msgSeq := g.networkManager.MsgSeq()
	stateMsg := &domain.GameMessage{
		MsgSeq:     msgSeq,
		SenderId:   g.GameSession.MyID(),
		ReceiverId: -1,
		Type: &domain.GameMessage_State{
			State: &domain.GameMessage_StateMsg{
				State: g.GameSession.State,
			},
		},
	}

	data, err := proto.Marshal(stateMsg)
	if err != nil {
		return err
	}
	err = g.sendToAllPlayers(&data, msgSeq)
	if err != nil {
		return err
	}
	return nil
}

func (g *Game) sendToAllPlayers(data *[]byte, msgSeq int64) error {
	for i := range g.GameSession.Players {
		if g.GameSession.Players[i].Player.Id == g.GameSession.MyID() {
			continue
		}
		ip := g.GameSession.Players[i].Player.IpAddress
		port := g.GameSession.Players[i].Player.Port
		dest := formatAddress(ip, port)
		err := g.networkManager.SendMsg(data, dest)
		if err != nil {
			return err
		}
		addr, err := network.StringToAddr(dest)
		if err != nil {
			return err
		}
		g.networkManager.NeedAck(network.NewMsg(*data, addr), msgSeq, true)
	}
	return nil
}

func formatAddress(ip string, port int32) string {
	return fmt.Sprintf("%s:%d", ip, port)
}

func (g *Game) sendSteer() error {
	msgSeq := g.networkManager.MsgSeq()
	steerMsg := &domain.GameMessage{
		SenderId:   g.GameSession.MyID(),
		ReceiverId: -1,
		MsgSeq:     msgSeq,
		Type: &domain.GameMessage_Steer{
			Steer: &domain.GameMessage_SteerMsg{
				Direction: g.controller.Direction(),
			},
		},
	}

	data, err := proto.Marshal(steerMsg)
	if err != nil {
		return err
	}
	err = g.networkManager.SendMsg(&data, g.GameSession.Node.MasterAddr())
	if err != nil {
		return err
	}
	addr, err := network.StringToAddr(g.GameSession.Node.MasterAddr())
	if err != nil {
		return err
	}
	g.networkManager.NeedAck(network.NewMsg(data, addr), msgSeq, true)
	return nil
}

func (g *Game) sendPing(addr string) error {
	msgSeq := g.networkManager.MsgSeq()
	pingMsg := &domain.GameMessage{
		SenderId:   g.GameSession.MyID(),
		ReceiverId: -1,
		MsgSeq:     msgSeq,
		Type: &domain.GameMessage_Ping{
			Ping: &domain.GameMessage_PingMsg{},
		},
	}

	data, err := proto.Marshal(pingMsg)
	if err != nil {
		return err
	}
	err = g.networkManager.SendMsg(&data, addr)
	if err != nil {
		return err
	}
	addrToAck, err := network.StringToAddr(addr)
	if err != nil {
		return err
	}
	g.networkManager.NeedAck(network.NewMsg(data, addrToAck), msgSeq, true)
	return nil
}

func (g *Game) sendRoleChangeMsg(addr string) error {
	msgSeq := g.networkManager.MsgSeq()
	chgMsg := &domain.GameMessage{
		SenderId:   g.GameSession.MyID(),
		ReceiverId: -1,
		MsgSeq:     msgSeq,
		Type: &domain.GameMessage_RoleChange{
			RoleChange: &domain.GameMessage_RoleChangeMsg{
				ReceiverRole: domain.NodeRole_DEPUTY,
			},
		},
	}

	data, err := proto.Marshal(chgMsg)
	if err != nil {
		return err
	}
	err = g.networkManager.SendMsg(&data, addr)
	if err != nil {
		return err
	}
	addrToAck, err := network.StringToAddr(addr)
	if err != nil {
		return err
	}
	g.networkManager.NeedAck(network.NewMsg(data, addrToAck), msgSeq, true)
	return nil
}

func GetIpAndPort(addr string) (string, int32) {
	split := strings.Split(addr, ":")
	if len(split) != 2 {
		return "", 0
	}
	port, err := strconv.Atoi(split[1])
	if err != nil {
		return "", 0
	}
	return split[0], int32(port)
}
