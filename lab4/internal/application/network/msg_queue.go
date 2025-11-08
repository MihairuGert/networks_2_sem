package network

import (
	"net"
	"sync"
)

type Msg struct {
	data   []byte
	addr   net.Addr
	msgNum int64
}

func (m Msg) Data() []byte {
	return m.data
}

func (m Msg) Addr() net.Addr {
	return m.addr
}

func (m Msg) MsgNum() int64 {
	return m.msgNum
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

	var temp []Msg
	copy(temp, mq.unreadMessages)
	mq.unreadMessages = mq.unreadMessages[:0]

	mq.sync.Unlock()
	return temp
}
