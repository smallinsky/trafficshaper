package bandwidth

import (
	"errors"
	"net"
	"sync"
)

const (
	defaultCapacity = 1024
)

var (
	ErrIncorrectLimit = errors.New("UL and DL limit should be greater than 0")
)

type Shaper interface {
	LimitDL(conn net.Conn, n int64)
	LimitUL(addr net.Conn, n int64)
	AddServerLimit(dl, ul int64) error
	AddCIDRLimit(cidr string, dl, ul int64) error

	GetDLCapaity(addr net.Conn) int64
	GetULCapaity(addr net.Conn) int64

	AddConn(conn net.Conn) error
	RemoveConn(conn net.Conn)
}

func NewShaper() Shaper {
	return &shaper{
		srv: &unlimtedBandwith,
		cidrs: CIDRLimit{
			limits:  make(map[string]cidrLimit),
			connThr: make(map[net.Conn]*BandThrot),
		},
	}
}

type shaper struct {
	srv   *BandThrot
	cidrs CIDRLimit
	mtx   sync.RWMutex
}

func (s *shaper) GetDLCapaity(conn net.Conn) int64 {
	l := s.cidrs.GetConnLimit(conn)
	if s.srv.dl.GetN() < l.dl.GetN() {
		return s.srv.dl.GetN()
	}
	return l.dl.GetN()
}

func (s *shaper) GetULCapaity(conn net.Conn) int64 {
	l := s.cidrs.GetConnLimit(conn)
	if s.srv.UL() < l.UL() {
		return s.srv.UL()
	}
	return l.UL()
}

func (s *shaper) LimitDL(conn net.Conn, n int64) {
	if s.srv != nil {
		s.srv.LimitDL(n)
	}
	s.cidrs.GetConnLimit(conn).LimitDL(n)
}

func (s *shaper) LimitUL(conn net.Conn, n int64) {
	if s.srv != nil {
		s.srv.LimitUL(n)
	}
	s.cidrs.GetConnLimit(conn).LimitUL(n)
}

func (s *shaper) AddServerLimit(dl, ul int64) error {
	if dl <= 0 || ul <= 0 {
		return ErrIncorrectLimit
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if s.srv != nil {
		s.srv = NewTBAThr(dl, ul)
	}
	s.srv.Reset(dl, ul)
	return nil
}

func (s *shaper) AddCIDRLimit(cidr string, dl, ul int64) error {
	if dl <= 0 || ul <= 0 {
		return ErrIncorrectLimit
	}
	return s.cidrs.AddLimit(cidr, dl, ul)
}

func (f *shaper) AddConn(conn net.Conn) error {
	return f.cidrs.AddConn(conn)
}

func (f *shaper) RemoveConn(conn net.Conn) {
	f.cidrs.RemoveConn(conn)
}
