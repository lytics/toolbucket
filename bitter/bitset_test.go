package bitter

import (
	"testing"

	"github.com/bmizerany/assert"
)

func TestBit(t *testing.T) {

	bmap := &Bit64Set{}
	bmap.Set(3)
	assert.T(t, bmap.Get(2) == false)
	assert.T(t, bmap.Get(3) == true)
	bmap.Set(5)
	assert.T(t, bmap.Get(7) == false)
	bmap.Set(7)
	assert.T(t, bmap.Get(7) == true)
	assert.T(t, bmap.Get(7) == true)
	bmap.Set(7)
	bmap.Set(7)
	bmap.Set(7)
	assert.T(t, bmap.Get(7) == true)
	assert.T(t, bmap.Get(3) == true)
	assert.T(t, bmap.Get(5) == true)
	assert.T(t, bmap.Get(0) == false)
	bmap.Clear(7)
	assert.T(t, bmap.Get(7) == false)
	assert.T(t, bmap.Get(7) == false)
	bmap.Clear(7)
	assert.T(t, bmap.Get(7) == false)

	//bounds tests
	bmap = &Bit64Set{}
	assert.T(t, bmap.Get(0) == false)
	bmap.Set(0)
	assert.T(t, bmap.Get(0) == true)
	assert.T(t, bmap.BitSet == uint64(1))
	assert.T(t, bmap.Get(63) == false)
	bmap.Set(63)
	assert.T(t, bmap.Get(63) == true)
	bmap.Clear(63)
	assert.T(t, bmap.Get(63) == false)
}

func TestBitLoop(t *testing.T) {
	bmap := &Bit64Set{}

	for i := 0; i < 64; i++ {
		bmap.Set(uint32(i))
	}
	assert.T(t, bmap.BitSet == ^uint64(0))
	for i := 63; i >= 0; i-- {
		bmap.Clear(uint32(i))
	}
	assert.T(t, bmap.BitSet == uint64(0))
}

func TestBitRange(t *testing.T) {
	b := &Bit64Set{}
	b.Set(0)
	b.Set(1)
	b.Set(2)
	b.Set(3)

	ex := uint64(0xf)
	bmap := &Bit64Set{}
	for i := 4; i < 65; i += 4 {
		bmap.SetRange(&Bit64RangeMask{uint32(i), uint32(i - 4), b.BitSet})
		//ex==0xf when i=4, 0xf0 when i=8, 0xf00 when i=12, etc
		assert.Equal(t, ex<<uint32(i-4), bmap.BitSet)
		bmap.SetRange(&Bit64RangeMask{uint32(i), uint32(i - 4), b.BitSet})
		assert.Equal(t, ex<<uint32(i-4), bmap.BitSet)
		bmap.ClearRange(uint32(i), uint32(i-4))
		assert.Equal(t, uint64(0), bmap.BitSet)
	}

	bmap.ClearRange(64, 0)
	assert.Equal(t, uint64(0), bmap.BitSet, bmap)
	bmap.SetRange(&Bit64RangeMask{64, 4, ^uint64(0)})
	bmap.SetRange(&Bit64RangeMask{64, 4, ^uint64(0)})
	bmap.SetRange(&Bit64RangeMask{64, 4, ^uint64(0)})
	assert.Equal(t, uint64(0xfffffffffffffff0), bmap.BitSet, bmap)
	bmap.ClearRange(64, 0)
	assert.Equal(t, uint64(0), bmap.BitSet)
	bmap.SetRange(&Bit64RangeMask{60, 4, ^uint64(0)})
	bmap.SetRange(&Bit64RangeMask{4, 0, ^uint64(0)})
	assert.Equal(t, uint64(0x0fffffffffffffff), bmap.BitSet, bmap)
	bmap.ClearRange(60, 4)
	bmap.Set(0)
	bmap.Set(1)
	bmap.Set(2)
	bmap.Set(3)
	bmap.Set(5)
	bmap.Set(63)
	bmap.Set(62)
	bmap.Set(61)
	bmap.Set(60)
	bmap.Set(59)
	bmap.Set(45)
	bmap.ClearRange(60, 4)
	assert.Equal(t, uint64(0xf00000000000000f), bmap.BitSet, bmap)
}
