package application

import (
	"errors"
	"fmt"
	"snake-game/internal/application/network"
	"snake-game/internal/domain"
	"time"

	"google.golang.org/protobuf/proto"
)

func (g *Game) startAnnouncement() error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Println("sending announcement")
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
		Config:   &g.GameSession.Config,
		CanJoin:  true,
		GameName: "asd",
	}

	announcementMsg := &domain.GameMessage{
		MsgSeq:     g.networkManager.MsgSeq,
		SenderId:   -1,
		ReceiverId: -1,
		Type: &domain.GameMessage_Announcement{
			Announcement: &domain.GameMessage_AnnouncementMsg{
				Games: []*domain.GameAnnouncement{&gameInfo},
			},
		},
	}

	g.networkManager.MsgSeq++

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
		MsgSeq:     g.networkManager.MsgSeq,
		SenderId:   -1,
		ReceiverId: -1,
		Type: &domain.GameMessage_Discover{
			Discover: &domain.GameMessage_DiscoverMsg{},
		},
	}

	g.networkManager.MsgSeq++

	data, err := proto.Marshal(discoverMsg)
	if err != nil {
		return err
	}

	err = g.networkManager.SendMsg(data, network.MulticastAddress)
	if err != nil {
		return err
	} else {
		fmt.Println("Discover message sent")
	}
	return nil
}

func (g *Game) startListening() {
	g.goroutinePool.Go(g.handleMessages)
	g.goroutinePool.Go(g.networkManager.ListenMulticast)
	g.goroutinePool.Go(g.networkManager.ListenUnicast)
}

func (g *Game) handleMessages() error {
	for msg := range g.handleChannel {
		var gameMsg domain.GameMessage
		if err := proto.Unmarshal(msg.Data(), &gameMsg); err != nil {
			fmt.Printf("Failed to unmarshal message: %v\n", err)
		}

		fmt.Printf("Received message type: %T from %s\n", gameMsg.Type, msg.Addr().String())

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
		}
	}
	return nil
}

func (g *Game) handleDiscover(msg *domain.GameMessage, addr string) error {
	if g.GameSession.Node.Role() == domain.NodeRole_MASTER {
		fmt.Printf("Received discover from %s, responding with announcement\n", addr)
		err := g.sendAnnouncementTo(addr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Game) handleAnnouncement(msg *domain.GameMessage, srcAddr string) error {
	announcement := msg.GetAnnouncement()
	if announcement == nil {
		return errors.New("invalid announcement")
	}

	fmt.Printf("Received announcement from %s with %d games\n",
		srcAddr, len(announcement.Games))

	for _, game := range announcement.Games {
		fmt.Printf("Game: %s, CanJoin: %v\n",
			game.GetGameName(), game.GetCanJoin())
	}
	return nil
}
