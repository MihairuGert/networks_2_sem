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
		SenderId:   -1,
		ReceiverId: -1,
		Type: &domain.GameMessage_Announcement{
			Announcement: &domain.GameMessage_AnnouncementMsg{
				Games: []*domain.GameAnnouncement{&gameInfo},
			},
		},
	}

	//g.networkManager.msgSeq++

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

func (g *Game) JoinGame(masterAddr string, gameName string, viewOnly bool) {
	joinMsg := &domain.GameMessage{
		SenderId:   -1,
		ReceiverId: -1,
		Type: &domain.GameMessage_Join{
			Join: &domain.GameMessage_JoinMsg{
				PlayerType: 0,
				GameName:   gameName,
			},
		},
	}

	//g.networkManager.msgSeq++

	data, err := proto.Marshal(joinMsg)
	if err != nil {
		fmt.Printf("Failed to marshal join: %v\n", err)
		return
	}

	err = g.networkManager.SendMsg(data, masterAddr)
	if err != nil {
		fmt.Printf("Failed to send join: %v\n", err)
	}
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
	if g.state != Connect {
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
	controller.IpAddress, controller.Port = GetIpAndPort(srcAddr)
	controller.SetPlayer(0, 0)
	g.addPlayer(&controller)
	return nil
}

func (g *Game) sendAckTo(originalMsg *domain.GameMessage, dest string) error {
	ackMsg := &domain.GameMessage{
		//MsgSeq:     g.networkManager.msgSeq,
		SenderId:   originalMsg.GetReceiverId(),
		ReceiverId: originalMsg.GetSenderId(),
		Type: &domain.GameMessage_Ack{
			Ack: &domain.GameMessage_AckMsg{},
		},
	}

	//g.networkManager.msgSeq++

	data, err := proto.Marshal(ackMsg)
	if err != nil {
		return err
	}
	err = g.networkManager.SendMsg(data, dest)
	if err != nil {
		return err
	}
	return nil
	// todo add connection control (sth like hashmap)
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
