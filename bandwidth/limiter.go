package bandwidth

import (
	"fmt"
	"time"

	"github.com/smallinsky/trafficshaper/throttle"
)

type throttler interface {
	WaitN(int64)
	GetN() int64
	Reset(int64)
}

var unlimtedBandwith = BandThrot{
	ul: throttle.NewNop(),
	dl: throttle.NewNop(),
}

func (b BandThrot) Unlimited() bool {
	return b == unlimtedBandwith
}
func NewTBAThr(dl, ul int64) *BandThrot {
	return &BandThrot{
		dl: throttle.NewTBA(time.Second, dl),
		ul: throttle.NewTBA(time.Second, ul),
	}
}

type BandThrot struct {
	ul throttler
	dl throttler
}

func (t *BandThrot) LimitUL(n int64) {
	t.ul.WaitN(n)
}

func (t *BandThrot) LimitDL(n int64) {
	t.dl.WaitN(n)
}

func (t *BandThrot) DL() int64 {
	return t.dl.GetN()
}

func (t *BandThrot) Reset(dl, ul int64) {
	t.dl.Reset(dl)
	t.ul.Reset(ul)
}

func (t *BandThrot) UL() int64 {
	return t.ul.GetN()
}

func (l *BandThrot) ToString() string {
	if l.ul != nil && l.dl != nil {
		return fmt.Sprintf("DL: %v  UL: %v", l.dl.GetN(), l.ul.GetN())
	}

	return fmt.Sprintf("DL: nil, UL: nil")
}
