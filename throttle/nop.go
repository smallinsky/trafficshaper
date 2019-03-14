package throttle

func NewNop() *Nop {
	return &Nop{}
}

type Nop struct {
}

func (o *Nop) WaitN(n int64) {
}

func (o *Nop) GetN() int64 {
	return 1<<63 - 1
}

func (o *Nop) Reset(n int64) {
}
