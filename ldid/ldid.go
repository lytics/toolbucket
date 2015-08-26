package ldid

import (
	"time"

	"github.com/lytics/toolbucket/bitter"
)

type LDIDMaker struct {
	Epoch     int64 //time converted to nano units.
	CreatorID uint16
	Counter   uint16

	windowtime int64
}

const tu = 1e7 // nsec, i.e. 10 msec

func (l *LDIDMaker) Next() *LDID {
	ctime := time.Now().UTC().UnixNano() / tu //truncate time to 10 msec
	ctime = ctime - l.Epoch
	if ctime != l.windowtime {
		l.windowtime = ctime
		l.Counter = 0
	}

	cnt := l.Counter

	l.Counter++
	if l.Counter == 0 {
		sleeptime :=
			time.Sleep()
	}

}

type LDID struct {
	CTime     time.Time
	CreatorID uint64
	Counter   uint64
}

func sleepTime(overtime int64) time.Duration {
	return time.Duration(overtime)*10*time.Millisecond -
		time.Duration(time.Now().UTC().UnixNano()%sonyflakeTimeUnit)*time.Nanosecond
}

func Marshal(ldid *LDID) uint64 {
	bf := &bitter.Bit64Set{}

}

func UnMarshal(f uint64, Epoch time.Time) *LDID {

}
