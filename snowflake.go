package snowflake

import (
	"errors"
	"time"
)

const (
	Uint10Mask = (uint64(1) << 10) - 1
	Uint12Mask = (uint64(1) << 12) - 1
	Uint41Mask = (uint64(1) << 41) - 1
)

// Clock abstract the standard time package
type Clock interface {
	Since(t time.Time) time.Duration
	Sleep(d time.Duration)
}

type defaultClock struct{}

func (defaultClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

func (defaultClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func DefaultClock() Clock {
	return defaultClock{}
}

// Options options for Snowflake
type Options struct {
	// Clock clock system, default to standard library
	Clock Clock
	// Epoch pre-defined zero time in Snowflake algorithm, required
	Epoch time.Time
	// ID unique unsigned integer indicate the ID of current Snowflake instance, maximum 10 bits wide, default to 0
	ID uint64
}

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
	chReq     chan interface{}
	chResp    chan uint64
	chStop    chan interface{}
	epoch     time.Time
	shiftedID uint64
	count     uint64
	clock     Clock
}

// New create a new instance of Snowflake
func New(opts Options) (Snowflake, error) {
	if opts.Clock == nil {
		opts.Clock = DefaultClock()
	}
	if opts.Epoch.IsZero() {
		return nil, errors.New("failed to create Snowflake: missing Epoch")
	}
	if opts.ID&Uint10Mask != opts.ID {
		return nil, errors.New("failed to create Snowflake: invalid ID")
	}
	sf := &snowflake{
		chReq:     make(chan interface{}),
		chResp:    make(chan uint64),
		chStop:    make(chan interface{}),
		epoch:     opts.Epoch,
		shiftedID: opts.ID << 12,
		clock:     opts.Clock,
	}
	go sf.run()
	return sf, nil
}

func (sf *snowflake) Stop() {
	close(sf.chStop)
}

func (sf *snowflake) run() {
	var nowT, lastT, seqID uint64
	for {
		select {
		case <-sf.chReq:
		retry:
			nowT = uint64(sf.clock.Since(sf.epoch) / time.Millisecond)
			if nowT == lastT {
				seqID = seqID + 1
				if seqID > Uint12Mask {
					sf.clock.Sleep(time.Millisecond)
					goto retry
				}
			} else {
				lastT = nowT
				seqID = 0
			}
			sf.count++
			sf.chResp <- ((nowT & Uint41Mask) << 22) | sf.shiftedID | seqID
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
