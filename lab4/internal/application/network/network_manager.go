package network

import (
	"net"
)

type NetworkManager struct {
	MulticastSocket *net.UDPConn
}

func NewNetworkManager() *NetworkManager {
	mcs, err := initMulticastSocket()
	if err != nil {
		panic(err)
	}
	nw := &NetworkManager{
		MulticastSocket: mcs,
	}
	return nw
}

func initMulticastSocket() (*net.UDPConn, error) {
	multicastAddr := "239.192.0.4:9192"

	addr, err := net.ResolveUDPAddr("udp", multicastAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (nm *NetworkManager) SendMsg(conn *net.UDPConn, msg []byte) error {
	_, err := conn.Write(msg)
	return err
}
