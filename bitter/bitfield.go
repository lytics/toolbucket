package bitter

import "bytes"

type Bit64Set struct {
	BitSet uint64
}

func (b *Bit64Set) Set(bit uint32) {
	b.BitSet |= (1 << bit)
}

func (b *Bit64Set) Clear(bit uint32) {
	b.BitSet &= ^(1 << bit)
}

func (b *Bit64Set) Get(bit uint32) bool {
	if b.BitSet&(1<<bit) == 0 {
		return false
	} else {
		return true
	}
}

func (b *Bit64Set) String() string {
	buf := bytes.Buffer{}
	for i := 63; i >= 0; i-- {
		if b.BitSet&(1<<uint32(i)) > 0 {
			buf.WriteString("1")
		} else {
			buf.WriteString("0")
		}
	}
	return buf.String()
}

func (b *Bit64Set) SetRange(r *Bit64RangeMask) {
	frontshift := (64 - r.Size())
	mask := (r.Mask << frontshift) >> frontshift
	mask = (mask << r.End)
	b.BitSet &= ^(mask)
	b.BitSet |= mask
}

func (b *Bit64Set) ClearRange(start, end uint32) {
	size := start - end
	shift := 64 - size
	mask := (^uint64(0)) << shift
	mask = mask >> shift
	b.BitSet &= ^(mask << (start - size))
}

type Bit64RangeMask struct {
	Start uint32
	End   uint32
	Mask  uint64
}

func (r *Bit64RangeMask) Size() uint32 {
	return r.Start - r.End
}
