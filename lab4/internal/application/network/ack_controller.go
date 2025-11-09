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
)

type AckStatus struct {
	wasAck      AckState
	doAutoCheck bool
	sendTime    time.Time

	msg    *Msg
	ackMsg *domain.GameMessage
}
type AckController struct {
	ackMap   map[int64]*AckStatus
	ackMutex sync.Mutex

	resendInterval time.Duration
	sendChan       *chan Msg
}

func NewAckController(sendChan *chan Msg) *AckController {
	ackMap := make(map[int64]*AckStatus)
	ac := &AckController{ackMap, sync.Mutex{}, 1000 * time.Millisecond, sendChan}
	return ac
}

func (ac *AckController) addAckMsg(msg *Msg, msgSec int64, doAutoCheck bool) {
	ac.ackMutex.Lock()
	defer ac.ackMutex.Unlock()
	ac.ackMap[msgSec] = &AckStatus{Sent, doAutoCheck, time.Now(), msg, nil}
}

func (ac *AckController) checkAck(msgNum int64) (bool, *domain.GameMessage) {
	ac.ackMutex.Lock()
	defer ac.ackMutex.Unlock()
	status, ok := ac.ackMap[msgNum]
	if ok && status.wasAck == Ack {
		delete(ac.ackMap, msgNum)
	}
	return ok && status.wasAck == Ack, status.ackMsg
}

func (ac *AckController) setAck(msgNum int64, ackMsg *domain.GameMessage) {
	ac.ackMutex.Lock()
	defer ac.ackMutex.Unlock()
	if _, ok := ac.ackMap[msgNum]; ok {
		ac.ackMap[msgNum].wasAck = Ack
		ac.ackMap[msgNum].ackMsg = ackMsg
	}
}

func (ac *AckController) daemonRoutine() {
	for {
		time.Sleep(ac.resendInterval)
		ac.ackMutex.Lock()

		var keysToDelete []int64
		var messagesToResend []*Msg

		for k, v := range ac.ackMap {
			if v.wasAck == Sent && time.Since(v.sendTime) > ac.resendInterval {
				ac.ackMap[k].sendTime = time.Now()
				messagesToResend = append(messagesToResend, v.msg)
			}

			if v.doAutoCheck && v.wasAck == Ack {
				keysToDelete = append(keysToDelete, k)
			}
		}

		for _, k := range keysToDelete {
			delete(ac.ackMap, k)
		}

		ac.ackMutex.Unlock()
		for _, msg := range messagesToResend {
			*ac.sendChan <- *msg
		}
	}
}
