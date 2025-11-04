package network

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"
)

const (
	MulticastAddress = "239.192.0.4:9192"
)

type Msg struct {
	data []byte
	addr net.Addr
}

func (m Msg) Data() []byte {
	return m.data
}

func (m Msg) Addr() net.Addr {
	return m.addr
}

type Manager struct {
	multicastSocket *net.UDPConn
	unicastSocket   *net.UDPConn
	MsgSeq          int64

	handleChannel *chan Msg
}

func NewNetworkManager(handleChannel *chan Msg) *Manager {
	mcs, err := newMulticastSocket()
	if err != nil {
		panic(err)
	}

	ucs, err := newUnicastSocket()
	if err != nil {
		panic(err)
	}

	nw := &Manager{
		multicastSocket: mcs,
		unicastSocket:   ucs,
		MsgSeq:          0,
		handleChannel:   handleChannel,
	}
	return nw
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

func (nm *Manager) SendMsg(msg []byte, addr string) error {
	temp := strings.Split(addr, ":")
	if len(temp) != 2 {
		return syscall.EINVAL
	}

	ip := net.ParseIP(temp[0])
	port, err := strconv.ParseInt(temp[1], 10, 0)
	if err != nil {
		return syscall.EINVAL
	}

	_, err = nm.unicastSocket.WriteTo(msg, &net.UDPAddr{IP: ip, Port: int(port)})
	return err
}

func (nm *Manager) ListenMulticast() error {
	buffer := make([]byte, 65507)

	for {
		n, srcAddr, err := nm.multicastSocket.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Multicast read error: %v\n", err)
			continue
		}

		*nm.handleChannel <- Msg{buffer[:n], srcAddr}
	}
}

func (nm *Manager) ListenUnicast() error {
	buffer := make([]byte, 65507)

	for {
		n, srcAddr, err := nm.unicastSocket.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Unicast read error: %v\n", err)
			continue
		}

		*nm.handleChannel <- Msg{buffer[:n], srcAddr}
	}
}
