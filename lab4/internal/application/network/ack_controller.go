package network

import (
	"snake-game/internal/domain"
	"sync"
	"time"
)

type AckState int32

const (
	Sent AckState = iota
	Ack
	StrongAck
)

type AckStatus struct {
	wasAck   AckState
	sendTime time.Time

	msg *domain.GameMessage
}
type AckController struct {
	ackMap   map[int64]AckStatus
	ackMutex sync.Mutex

	resendInterval time.Duration
}

func NewAckController(resendInt time.Duration) *AckController {
	ackMap := make(map[int64]AckStatus)
	ac := &AckController{ackMap, sync.Mutex{}, resendInt}
	go ac.daemonRoutine()
	return ac
}

func (ac *AckController) addAckMsg(msg *domain.GameMessage) {
	ac.ackMutex.Lock()
	defer ac.ackMutex.Unlock()
	ac.ackMap[msg.MsgSeq] = AckStatus{Sent, time.Now(), msg}
}

func (ac *AckController) checkAck(msgNum int64) bool {
	ac.ackMutex.Lock()
	defer ac.ackMutex.Unlock()
	status, ok := ac.ackMap[msgNum]
	delete(ac.ackMap, msgNum)
	return ok && (status.wasAck == StrongAck || status.wasAck == Ack)
}

func (ac *AckController) daemonRoutine() {
	for {
		ac.ackMutex.Lock()
		for k, v := range ac.ackMap {
			if v.wasAck == StrongAck {
				continue
			}
			if v.wasAck == Ack {
				delete(ac.ackMap, k)
				continue
			}
			if time.Since(v.sendTime) > ac.resendInterval {
				v.sendTime = time.Now()

			}
		}
		ac.ackMutex.Unlock()
		time.Sleep(ac.resendInterval / 10)
	}
}
