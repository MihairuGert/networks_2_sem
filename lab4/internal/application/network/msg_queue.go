package network

import (
	"net"
	"sync"
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

type MsgQueue struct {
	unreadMessages []Msg
	sync           sync.Mutex
}

func NewMsgQueue() *MsgQueue {
	return &MsgQueue{sync: sync.Mutex{}}
}

func (mq *MsgQueue) addMsg(msg Msg) {
	mq.sync.Lock()
	mq.unreadMessages = append(mq.unreadMessages, msg)
	mq.sync.Unlock()
}

func (mq *MsgQueue) readAllMsg() []Msg {
	mq.sync.Lock()
	defer mq.sync.Unlock()

	temp := mq.unreadMessages
	mq.unreadMessages = mq.unreadMessages[:0]
	return temp
}
