package snowflake

import (
	"time"
)

const (
	uint10Mask = (uint64(1) << 10) - 1
	uint12Mask = (uint64(1) << 12) - 1
	uint41Mask = (uint64(1) << 41) - 1
)

// Snowflake the main interface
type Snowflake interface {
	// Stop stop the snowflake, release all related resources
	// can not stop twice, NewID() invocation will panic after stopped
	Stop()

	// Count returns the count of generated ids
	Count() uint64

	// NewID returns a new id
	NewID() uint64
}

type snowflake struct {
	chReq        chan interface{}
	chResp       chan uint64
	chStop       chan interface{}
	startTime    time.Time
	shiftedInsID uint64
	count        uint64
}

// New create a new instance of Snowflake
// startTime, the zero time for snowflake algorithm
// instanceId, should be a unique unsigned integer with maximum 10 bits
func New(startTime time.Time, instanceId uint64) Snowflake {
	sf := &snowflake{
		chReq:        make(chan interface{}),
		chResp:       make(chan uint64),
		chStop:       make(chan interface{}),
		startTime:    startTime,
		shiftedInsID: (instanceId & uint10Mask) << 12,
	}
	go sf.run()
	return sf
}

func (sf *snowflake) Stop() {
	close(sf.chStop)
}

func (sf *snowflake) run() {
	var lastT uint64
	var seqID uint64
	for {
		select {
		case <-sf.chReq:
		retry:
			nowT := uint64(time.Since(sf.startTime) / time.Millisecond)
			if nowT == lastT {
				seqID = seqID + 1
				if seqID > uint12Mask {
					time.Sleep(time.Millisecond)
					goto retry
				}
			} else {
				lastT = nowT
				seqID = 0
			}
			sf.count++
			sf.chResp <- ((nowT & uint41Mask) << 22) | sf.shiftedInsID | seqID
		case <-sf.chStop:
			return
		}
	}
}

func (sf *snowflake) Count() uint64 {
	return sf.count
}

func (sf *snowflake) NewID() uint64 {
	select {
	case sf.chReq <- nil:
		return <-sf.chResp
	case <-sf.chStop:
		panic("NewID() invoked after snowflake stopped")
	}
}
