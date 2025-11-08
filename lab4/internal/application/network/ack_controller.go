package network

import (
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

	msg *Msg
}
type AckController struct {
	ackMap   map[int64]*AckStatus
	ackMutex sync.Mutex

	resendInterval time.Duration
	sendChan       *chan Msg
}

func NewAckController(sendChan *chan Msg, resendInt time.Duration) *AckController {
	ackMap := make(map[int64]*AckStatus)
	ac := &AckController{ackMap, sync.Mutex{}, resendInt, sendChan}
	go ac.daemonRoutine()
	return ac
}

func (ac *AckController) addAckMsg(msg *Msg, msgSec int64, doAutoCheck bool) {
	ac.ackMutex.Lock()
	defer ac.ackMutex.Unlock()
	ac.ackMap[msgSec] = &AckStatus{Sent, doAutoCheck, time.Now(), msg}
}

func (ac *AckController) checkAck(msgNum int64) bool {
	ac.ackMutex.Lock()
	defer ac.ackMutex.Unlock()
	status, ok := ac.ackMap[msgNum]
	if ok && status.wasAck == Ack {
		delete(ac.ackMap, msgNum)
	}
	return ok && status.wasAck == Ack
}

func (ac *AckController) setAck(msgNum int64) {
	ac.ackMutex.Lock()
	defer ac.ackMutex.Unlock()
	if ack, ok := ac.ackMap[msgNum]; ok {
		ack.wasAck = Ack
	}
}

func (ac *AckController) daemonRoutine() {
	for {
		ac.ackMutex.Lock()

		var keysToDelete []int64
		var messagesToResend []*Msg

		for k, v := range ac.ackMap {
			if time.Since(v.sendTime) > ac.resendInterval {
				v.sendTime = time.Now()
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
		time.Sleep(ac.resendInterval)
	}
}
