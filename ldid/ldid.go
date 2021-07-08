package ldid

import "time"

type LDIDMaker struct {
	Epoch     int64 //time converted to nano units.
	CreatorID uint16
	Counter   uint16

	windowtime int64
}

const tu = int64(10 * time.Millisecond) // 10 msec

func (l *LDIDMaker) Next() *LDID {
	ctime := now10milli() //truncate time to 10 msec
	ctime = ctime - l.Epoch
	if ctime != l.windowtime {
		l.windowtime = ctime
		l.Counter = 0
	}

	//cnt := l.Counter

	l.Counter++
	if l.Counter == 0 {
		time.Sleep(sleeptime(ctime))
	}

	return nil
}

type LDID struct {
	CTime     time.Time
	CreatorID uint64
	Counter   uint64
}

func Marshal(ldid *LDID) uint64 {
	//bf := &bitter.Bit64Set{}
	return 0
}

func UnMarshal(f uint64, Epoch time.Time) *LDID {
	return nil
}

func sleeptime(ctime int64) time.Duration {
	stime := ctime * tu
	stime = stime - (nownano() % int64(tu))
	return dur(stime)
}

func dur(t int64) time.Duration {
	return time.Duration(t)
}

func now10milli() int64 {
	return nownano() / tu
}

func nownano() int64 {
	return time.Now().UTC().UnixNano()
}
