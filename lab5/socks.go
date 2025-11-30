package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/miekg/dns"
	"golang.org/x/sys/unix"
)

const (
	socksVersion5 = 0x05
	cmdConnect    = 0x01
	atypIP4       = 0x01
	atypDomain    = 0x03
	atypIP6       = 0x04
)

type stage int

const (
	auth stage = iota
	request
	establish
)

type Proxy struct {
	listener *net.TCPListener
	conns    map[int]*ClientConn
	dnsConn  *net.UDPConn
	dnsMap   map[uint16]*ClientConn
}

type ClientConn struct {
	clientFd    int
	clientConn  *net.TCPConn
	remoteFd    int
	remoteConn  *net.TCPConn
	stage       stage
	buffer      []byte
	readOffset  int
	writeOffset int
	targetHost  string
	targetPort  uint16
	dnsQueryID  uint16
}

func NewProxy(port int) (*Proxy, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	dnsConn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(8, 8, 8, 8),
		Port: 53,
	})
	if err != nil {
		return nil, err
	}

	return &Proxy{
		listener: listener,
		conns:    make(map[int]*ClientConn),
		dnsConn:  dnsConn,
		dnsMap:   make(map[uint16]*ClientConn),
	}, nil
}

func (p *Proxy) Run() error {
	listenerFd, err := p.getFdFromConn(p.listener)
	if err != nil {
		return err
	}
	defer unix.Close(listenerFd)

	dnsFd, err := p.getFdFromConn(p.dnsConn)
	if err != nil {
		return err
	}
	defer unix.Close(dnsFd)

	epollFd, err := unix.EpollCreate1(0)
	if err != nil {
		return err
	}
	defer unix.Close(epollFd)

	if err := unix.EpollCtl(epollFd, unix.EPOLL_CTL_ADD, listenerFd, &unix.EpollEvent{
		Events: unix.EPOLLIN,
		Fd:     int32(listenerFd),
	}); err != nil {
		return err
	}

	if err := unix.EpollCtl(epollFd, unix.EPOLL_CTL_ADD, dnsFd, &unix.EpollEvent{
		Events: unix.EPOLLIN,
		Fd:     int32(dnsFd),
	}); err != nil {
		return err
	}

	events := make([]unix.EpollEvent, 64)
	for {
		n, err := unix.EpollWait(epollFd, events, -1)
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}
			return err
		}

		for i := 0; i < n; i++ {
			fd := int(events[i].Fd)

			switch {
			case fd == listenerFd:
				if err := p.acceptClient(epollFd); err != nil {
					log.Printf("Accept error: %v", err)
				}
			case fd == dnsFd:
				if err := p.handleDNSResponse(); err != nil {
					log.Printf("DNS error: %v", err)
				}
			default:
				if err := p.handleClientData(fd, epollFd, events[i].Events); err != nil {
					log.Printf("Client handling error: %v", err)
					p.closeClient(fd)
				}
			}
		}
	}
}

func (p *Proxy) getFdFromConn(conn interface{}) (int, error) {
	switch c := conn.(type) {
	case *net.TCPListener:
		f, err := c.File()
		if err != nil {
			return -1, err
		}
		return int(f.Fd()), nil
	case *net.UDPConn:
		f, err := c.File()
		if err != nil {
			return -1, err
		}
		return int(f.Fd()), nil
	case *net.TCPConn:
		f, err := c.File()
		if err != nil {
			return -1, err
		}
		return int(f.Fd()), nil
	default:
		return -1, errors.New("unsupported connection type")
	}
}

func (p *Proxy) acceptClient(epollFd int) error {
	clientConn, err := p.listener.AcceptTCP()
	if err != nil {
		return err
	}

	clientFd, err := p.getFdFromConn(clientConn)
	if err != nil {
		clientConn.Close()
		return err
	}

	if err := unix.SetNonblock(clientFd, true); err != nil {
		clientConn.Close()
		unix.Close(clientFd)
		return err
	}

	if err := unix.EpollCtl(epollFd, unix.EPOLL_CTL_ADD, clientFd, &unix.EpollEvent{
		Events: unix.EPOLLIN | unix.EPOLLET,
		Fd:     int32(clientFd),
	}); err != nil {
		clientConn.Close()
		unix.Close(clientFd)
		return err
	}

	p.conns[clientFd] = &ClientConn{
		clientFd:   clientFd,
		clientConn: clientConn,
		stage:      auth,
		buffer:     make([]byte, 4096),
	}

	log.Printf("New client connected: %d", clientFd)
	return nil
}

func (p *Proxy) handleClientData(fd int, epollFd int, events uint32) error {
	client, ok := p.conns[fd]
	if !ok {
		return fmt.Errorf("unknown client: %d", fd)
	}

	if events&unix.EPOLLIN != 0 {
		if err := p.readFromClient(client); err != nil {
			return err
		}
	}

	if events&unix.EPOLLOUT != 0 {
		// gonna use it if decide to rewrite without the best of the best goroutines.
	}

	if events&(unix.EPOLLHUP|unix.EPOLLERR) != 0 {
		return errors.New("connection error")
	}

	return nil
}

func (p *Proxy) readFromClient(client *ClientConn) error {
	for {
		n, err := unix.Read(client.clientFd, client.buffer[client.readOffset:])
		if err != nil {
			if errors.Is(err, unix.EAGAIN) {
				break
			}
			return err
		}
		if n == 0 {
			return errors.New("client disconnected")
		}

		client.readOffset += n

		switch client.stage {
		case auth:
			if err := p.handleAuth(client); err != nil {
				return err
			}
		case request:
			if err := p.handleRequest(client); err != nil {
				return err
			}
		case establish:
			if client.remoteConn != nil {
				if _, err := client.remoteConn.Write(client.buffer[:client.readOffset]); err != nil {
					return err
				}
				client.readOffset = 0
			}
		}
	}
	return nil
}

func (p *Proxy) handleAuth(client *ClientConn) error {
	if client.readOffset < 2 {
		return nil
	}

	if client.buffer[0] != socksVersion5 {
		return fmt.Errorf("unsupported SOCKS version: %d", client.buffer[0])
	}

	nMethods := int(client.buffer[1])
	if client.readOffset < 2+nMethods {
		return nil
	}

	// So we do not support any auth so far.
	noAuthSupported := false
	for i := 0; i < nMethods; i++ {
		if client.buffer[2+i] == 0x00 {
			noAuthSupported = true
			break
		}
	}

	if !noAuthSupported {
		response := []byte{socksVersion5, 0xFF}
		if _, err := unix.Write(client.clientFd, response); err != nil {
			return err
		}
		return errors.New("no supported auth methods")
	}

	response := []byte{socksVersion5, 0x00}
	if _, err := unix.Write(client.clientFd, response); err != nil {
		return err
	}

	client.stage = request
	client.readOffset = 0
	log.Printf("Client %d authenticated", client.clientFd)
	return nil
}

func (p *Proxy) handleRequest(client *ClientConn) error {
	if client.readOffset < 4 {
		return nil
	}

	if client.buffer[0] != socksVersion5 {
		return fmt.Errorf("invalid SOCKS version in request: %d", client.buffer[0])
	}
	if client.buffer[1] != cmdConnect {
		return fmt.Errorf("unsupported command: %d", client.buffer[1])
	}

	aTyp := client.buffer[3]
	var host string

	switch aTyp {
	case atypIP4:
		if client.readOffset < 10 {
			return nil
		}
		host = net.IP(client.buffer[4:8]).String()
		client.targetPort = binary.BigEndian.Uint16(client.buffer[8:10])
	case atypDomain:
		if client.readOffset < 5 {
			return nil
		}
		domainLen := int(client.buffer[4])
		if client.readOffset < 7+domainLen {
			return nil
		}
		host = string(client.buffer[5 : 5+domainLen])
		client.targetPort = binary.BigEndian.Uint16(client.buffer[5+domainLen : 7+domainLen])
	case atypIP6:
		if client.readOffset < 22 {
			return nil
		}
		host = net.IP(client.buffer[4:20]).String()
		client.targetPort = binary.BigEndian.Uint16(client.buffer[20:22])
	default:
		return fmt.Errorf("unsupported address type: %d", aTyp)
	}

	client.targetHost = host
	client.readOffset = 0

	log.Printf("Client %d requesting connection to %s:%d", client.clientFd, host, client.targetPort)

	if aTyp == atypDomain {
		return p.resolveHost(client, host)
	}
	return p.connectToRemote(client, host)
}

func (p *Proxy) resolveHost(client *ClientConn, host string) error {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(host), dns.TypeA)
	msg.RecursionDesired = true

	client.dnsQueryID = uint16(len(p.dnsMap) + 1)
	msg.Id = client.dnsQueryID
	p.dnsMap[client.dnsQueryID] = client

	rawMsg, err := msg.Pack()
	if err != nil {
		return err
	}

	_, err = p.dnsConn.Write(rawMsg)
	if err != nil {
		return err
	}

	log.Printf("DNS query sent for %s (ID: %d)", host, client.dnsQueryID)
	return nil
}

func (p *Proxy) handleDNSResponse() error {
	buf := make([]byte, 512)
	n, err := p.dnsConn.Read(buf)
	if err != nil {
		return err
	}

	msg := new(dns.Msg)
	if err := msg.Unpack(buf[:n]); err != nil {
		return err
	}

	client, ok := p.dnsMap[msg.Id]
	if !ok {
		return fmt.Errorf("unknown DNS query ID: %d", msg.Id)
	}
	delete(p.dnsMap, msg.Id)

	if msg.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("DNS resolution failed: %d", msg.Rcode)
	}

	var ip string
	for _, answer := range msg.Answer {
		if a, ok := answer.(*dns.A); ok {
			ip = a.A.String()
			break
		}
	}

	if ip == "" {
		return errors.New("no IP address found in DNS response")
	}

	log.Printf("DNS resolved %s -> %s", client.targetHost, ip)
	return p.connectToRemote(client, ip)
}

func (p *Proxy) connectToRemote(client *ClientConn, host string) error {
	targetAddr := fmt.Sprintf("%s:%d", host, client.targetPort)
	log.Printf("Connecting to %s", targetAddr)

	remoteConn, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		response := []byte{socksVersion5, 0x01, 0x00, 0x01, 0, 0, 0, 0, 0, 0}
		unix.Write(client.clientFd, response)
		return err
	}

	remoteTCP := remoteConn.(*net.TCPConn)
	remoteFd, err := p.getFdFromConn(remoteTCP)
	if err != nil {
		remoteTCP.Close()
		return err
	}

	if err := unix.SetNonblock(remoteFd, true); err != nil {
		remoteTCP.Close()
		unix.Close(remoteFd)
		return err
	}

	client.remoteConn = remoteTCP
	client.remoteFd = remoteFd

	response := make([]byte, 10)
	response[0] = socksVersion5
	response[1] = 0x00
	response[2] = 0x00
	response[3] = 0x01
	copy(response[4:8], net.IPv4(0, 0, 0, 0))
	binary.BigEndian.PutUint16(response[8:10], 0)

	if _, err := unix.Write(client.clientFd, response); err != nil {
		remoteTCP.Close()
		unix.Close(remoteFd)
		return err
	}

	client.stage = establish
	log.Printf("Connection established to %s:%d", host, client.targetPort)

	go p.relayData(client)

	return nil
}

func (p *Proxy) relayData(client *ClientConn) {
	defer p.closeClient(client.clientFd)

	buffer := make([]byte, 4096)
	for {
		n, err := client.remoteConn.Read(buffer)
		if err != nil {
			break
		}

		if _, err := unix.Write(client.clientFd, buffer[:n]); err != nil {
			break
		}
	}
}

func (p *Proxy) closeClient(fd int) {
	if client, ok := p.conns[fd]; ok {
		log.Printf("Closing client connection: %d", fd)
		delete(p.conns, fd)

		if client.clientConn != nil {
			client.clientConn.Close()
		}
		if client.remoteConn != nil {
			client.remoteConn.Close()
		}
		if client.dnsQueryID != 0 {
			delete(p.dnsMap, client.dnsQueryID)
		}
		unix.Close(fd)
		if client.remoteFd != 0 {
			unix.Close(client.remoteFd)
		}
	}
}

func main() {
	fmt.Print("Enter port: ")
	var port int
	_, err := fmt.Scan(&port)
	if err != nil {
		log.Fatal("Error reading port:", err)
	}

	proxy, err := NewProxy(port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("SOCKS5 proxy started on port %d", port)
	log.Fatal(proxy.Run())
}
