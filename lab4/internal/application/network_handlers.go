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
		err2 := g.sendAnnouncementTo(network.MulticastAddress)
		if err2 != nil {
			return err2
		}
	}
	return nil
}

func (g *Game) sendAnnouncementTo(addr string) error {
	gameInfo := domain.GameAnnouncement{
		Players:  g.GameSession.State.Players,
		Config:   g.GameSession.Config,
		CanJoin:  true,
		GameName: "asd",
	}

	announcementMsg := &domain.GameMessage{
		MsgSeq:     g.networkManager.MsgSeq(),
		SenderId:   -1,
		ReceiverId: -1,
		Type: &domain.GameMessage_Announcement{
			Announcement: &domain.GameMessage_AnnouncementMsg{
				Games: []*domain.GameAnnouncement{&gameInfo},
			},
		},
	}

	data, err := proto.Marshal(announcementMsg)
	if err != nil {
		return err
	}
	err = g.networkManager.SendMsg(data, addr)
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

	err = g.networkManager.SendMsg(data, network.MulticastAddress)
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

	err = g.networkManager.SendMsg(data, masterAddr)
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
			return err
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
			g.networkManager.SetAck(gameMsg.MsgSeq)
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

	err := g.sendAckTo(msg, srcAddr)
	if err != nil {
		return err
	}
	controller := network.Controller{}
	ipAddress, port := GetIpAndPort(srcAddr)
	controller.SetIpAndPort(ipAddress, port)
	id := g.GameSession.GetFreePlayerId()
	controller.SetId(int32(id))
	controller.SetPlayer(0, 0, msg.GetJoin().PlayerName, int32(id))
	g.addPlayer(&controller)
	return nil
}

func (g *Game) sendAckTo(originalMsg *domain.GameMessage, dest string) error {
	ackMsg := &domain.GameMessage{
		MsgSeq:     g.networkManager.MsgSeq(),
		SenderId:   originalMsg.GetReceiverId(),
		ReceiverId: originalMsg.GetSenderId(),
		Type: &domain.GameMessage_Ack{
			Ack: &domain.GameMessage_AckMsg{},
		},
	}

	data, err := proto.Marshal(ackMsg)
	if err != nil {
		return err
	}
	err = g.networkManager.SendMsg(data, dest)
	if err != nil {
		return err
	}
	return nil
}

func (g *Game) sendState() error {
	stateMsg := &domain.GameMessage{
		MsgSeq:     g.networkManager.MsgSeq(),
		SenderId:   -1,
		ReceiverId: -1,
		Type: &domain.GameMessage_State{
			State: &domain.GameMessage_StateMsg{
				State: &g.GameSession.State,
			},
		},
	}

	data, err := proto.Marshal(stateMsg)
	if err != nil {
		return err
	}
	err = g.networkManager.SendMsg(data, network.MulticastAddress)
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
