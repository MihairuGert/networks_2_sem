package network

import (
	"fmt"
	"net"
	"snake-game/internal/domain"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"google.golang.org/protobuf/proto"
)

const (
	MulticastAddress = "239.192.0.4:9192"
)

type Manager struct {
	multicastSocket *net.UDPConn
	unicastSocket   *net.UDPConn

	stateDelayMs time.Duration

	shouldStop *bool

	msgQueue    *MsgQueue
	msgSeq      int64
	msqSeqMutex sync.Mutex

	sendChan chan Msg

	ackController *AckController

	sendPingMap sync.Map
	recvPingMap sync.Map
}

func (nm *Manager) GetAddr() string {
	localAddr := nm.unicastSocket.LocalAddr().(*net.UDPAddr)
	return fmt.Sprintf("%s:%d", getOutboundIP(), localAddr.Port)
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func (nm *Manager) Close() {
	nm.multicastSocket.Close()
	nm.unicastSocket.Close()
}

func (nm *Manager) SetErr(seqNum int64, errMsg *domain.GameMessage) {
	nm.ackController.setErr(seqNum, errMsg)
}

func (nm *Manager) SetAck(seqNum int64, ackMsg *domain.GameMessage) {
	nm.ackController.setAck(seqNum, ackMsg)
}

func (nm *Manager) NeedAck(msg *Msg, seqNum int64, doAutoCheck bool) {
	nm.ackController.addAckMsg(msg, seqNum, doAutoCheck)
}

func (nm *Manager) CheckAck(seqNum int64) (bool, *domain.GameMessage) {
	return nm.ackController.checkAck(seqNum)
}

func (nm *Manager) MsgSeq() int64 {
	nm.msqSeqMutex.Lock()
	defer nm.msqSeqMutex.Unlock()
	nm.msgSeq++
	return nm.msgSeq - 1
}

func NewNetworkManager(shouldStop *bool) *Manager {
	mcs, err := newMulticastSocket()
	if err != nil {
		panic(err)
	}

	ucs, err := newUnicastSocket()
	if err != nil {
		panic(err)
	}

	mq := NewMsgQueue()

	sendChan := make(chan Msg, 100)

	ac := NewAckController(&sendChan, shouldStop)

	nw := &Manager{
		multicastSocket: mcs,
		unicastSocket:   ucs,
		msgSeq:          0,
		msgQueue:        mq,
		ackController:   ac,
		sendChan:        sendChan,
		msqSeqMutex:     sync.Mutex{},
		shouldStop:      shouldStop,
		sendPingMap:     sync.Map{},
		recvPingMap:     sync.Map{},
	}
	go nw.sendGoroutine()
	return nw
}

func (nm *Manager) StartAckDaemonWithDuration(duration time.Duration) {
	nm.ackController.resendInterval = duration
	go nm.ackController.daemonRoutine()
}

func newUnicastSocket() (*net.UDPConn, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		err := conn.Close()
		if err != nil {
			return nil, err
		}
		return nil, err
	}
	return conn, nil
}

func newMulticastSocket() (*net.UDPConn, error) {
	multicastAddr, err := net.ResolveUDPAddr("udp", MulticastAddress)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenMulticastUDP("udp", nil, multicastAddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (nm *Manager) SendMsg(msg *[]byte, strAddr string) error {
	if strAddr != MulticastAddress {
		nm.sendPingMap.Store(strAddr, time.Now())
	}

	addr, err := StringToAddr(strAddr)
	if err != nil {
		return err
	}
	nm.sendChan <- Msg{
		data: *msg,
		addr: addr,
	}
	return nil
}

func StringToAddr(addr string) (*net.UDPAddr, error) {
	temp := strings.Split(addr, ":")
	if len(temp) != 2 {
		return nil, syscall.EINVAL
	}

	ip := net.ParseIP(temp[0])
	port, err := strconv.ParseInt(temp[1], 10, 0)
	if err != nil {
		return nil, syscall.EINVAL
	}
	return &net.UDPAddr{
		IP:   ip,
		Port: int(port),
	}, nil
}

func (nm *Manager) ListenMulticast() error {
	for {
		if *nm.shouldStop {
			return nil
		}
		buffer := make([]byte, 65507)
		n, srcAddr, err := nm.multicastSocket.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		nm.msgQueue.addMsg(Msg{buffer[:n], srcAddr})
	}
}

func (nm *Manager) ListenUnicast() error {
	for {
		if *nm.shouldStop {
			return nil
		}
		buffer := make([]byte, 65507)
		n, srcAddr, err := nm.unicastSocket.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		if srcAddr.String() == nm.unicastSocket.LocalAddr().String() {
			continue
		}

		var gameMsg domain.GameMessage
		if err := proto.Unmarshal(buffer[:n], &gameMsg); err == nil {
			switch gameMsg.Type.(type) {
			case *domain.GameMessage_Announcement, *domain.GameMessage_Discover, *domain.GameMessage_Ack, *domain.GameMessage_Join, *domain.GameMessage_RoleChange:
			default:
				ackMsg := &domain.GameMessage{
					MsgSeq:     gameMsg.MsgSeq,
					SenderId:   -1,
					ReceiverId: gameMsg.SenderId,
					Type: &domain.GameMessage_Ack{
						Ack: &domain.GameMessage_AckMsg{},
					},
				}
				data, err := proto.Marshal(ackMsg)
				if err == nil {
					nm.sendChan <- Msg{data: data, addr: srcAddr}
				}
			}
		}

		nm.msgQueue.addMsg(Msg{buffer[:n], srcAddr})
		nm.recvPingMap.Store(srcAddr.String(), time.Now())
	}
}

func (nm *Manager) sendGoroutine() {
	for {
		for msg := range nm.sendChan {
			if *nm.shouldStop {
				return
			}
			if msg.Addr().String() == nm.unicastSocket.LocalAddr().String() {
				continue
			}
			_, err := nm.unicastSocket.WriteTo(msg.data, msg.addr)
			if err != nil {
				return
			}
		}
	}
}

func (nm *Manager) GetUnreadMessages() []Msg {
	return nm.msgQueue.readAllMsg()
}

func (nm *Manager) GetWhoSentLessThan(duration time.Duration) []string {
	result := make([]string, 0)
	now := time.Now()

	nm.sendPingMap.Range(func(key, value interface{}) bool {
		addr, ok := key.(string)
		if !ok {
			return true
		}

		lastPingTime, ok := value.(time.Time)
		if !ok {
			return true
		}

		if now.Sub(lastPingTime) > duration {
			nm.recvPingMap.Delete(addr)
			result = append(result, addr)
		}

		return true
	})

	return result
}

func (nm *Manager) GetWhoRecvLessThan(duration time.Duration) []string {
	result := make([]string, 0)
	now := time.Now()

	nm.recvPingMap.Range(func(key, value interface{}) bool {
		addr, ok := key.(string)
		if !ok {
			return true
		}

		lastPingTime, ok := value.(time.Time)
		if !ok {
			return true
		}

		if now.Sub(lastPingTime) > duration {
			nm.recvPingMap.Delete(addr)
			result = append(result, addr)
		}

		return true
	})

	return result
}
