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
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if g.GameSession == nil || g.GameSession.Node.Role() != domain.NodeRole_MASTER {
			continue
		}
		err2 := g.sendAnnouncementTo(network.MulticastAddress)
		if err2 != nil {
			return err2
		}
	}
	return nil
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
	messages := g.networkManager.GetUnreadMessages()
	for _, msg := range messages {
		var gameMsg domain.GameMessage
		if err := proto.Unmarshal(msg.Data(), &gameMsg); err != nil {
			println("Failed to unmarshal gameMessage")
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
		}
	}
	return nil
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

	player, canJoin := g.GameSession.AddPlayer(&gp)

	if canJoin {
		msg.ReceiverId = id
		err := g.sendAckTo(msg, srcAddr)
		g.Renderer.AddPlayer(player.Player.Name, GetRoleString(player.Player.Role), int(player.Player.Score))
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
	return nil
}

func (g *Game) sendState() error {
	stateMsg := &domain.GameMessage{
		MsgSeq:     g.networkManager.MsgSeq(),
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
	err = g.sendToAllPlayers(&data)
	if err != nil {
		return err
	}
	return nil
}

func (g *Game) sendToAllPlayers(data *[]byte) error {
	for i := range g.GameSession.Players {
		if g.GameSession.Players[i].Player.Id == g.GameSession.MyID() {
			continue
		}
		ip := g.GameSession.Players[i].Player.IpAddress
		port := g.GameSession.Players[i].Player.Port
		err := g.networkManager.SendMsg(data, ip+":"+strconv.FormatInt(int64(port), 10))
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Game) sendSteer() error {
	steerMsg := &domain.GameMessage{
		SenderId:   g.GameSession.MyID(),
		ReceiverId: -1,
		MsgSeq:     g.networkManager.MsgSeq(),
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
