package bandwidth

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

var (
	ErrFailedToGetNetIP = errors.New("failed to obtain net.IP from net.Addr type")
)

type cidrLimit struct {
	netIP *net.IPNet
	dl    int64
	ul    int64
}

func NewCIDRLimit() *CIDRLimit {
	return &CIDRLimit{
		limits:  make(map[string]cidrLimit),
		connThr: make(map[net.Conn]*BandThrot),
	}
}

type CIDRLimit struct {
	mtx     sync.RWMutex
	limits  map[string]cidrLimit
	connThr map[net.Conn]*BandThrot
}

func (f *CIDRLimit) AddLimit(cidr string, dl, ul int64) error {
	_, ipv4Net, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("failed to parse CIDR(%v)", ErrFailedToGetNetIP)
	}
	k := ipv4Net.String()

	f.limits[k] = cidrLimit{
		netIP: ipv4Net,
		dl:    dl,
		ul:    ul,
	}

	for k, _ := range f.connThr {
		a, ok := k.RemoteAddr().(*net.TCPAddr)
		if !ok {
			return ErrFailedToGetNetIP
		}
		if ipv4Net.Contains(a.IP) {
			f.AddConn(k)
		}
	}
	return nil
}

func (f *CIDRLimit) AddConn(conn net.Conn) error {
	a, ok := conn.RemoteAddr().(*net.TCPAddr)
	if !ok {
		return ErrFailedToGetNetIP
	}

	f.mtx.Lock()
	defer f.mtx.Unlock()

	found := false
	for _, v := range f.limits {
		if v.netIP.Contains(a.IP) {
			f.connThr[conn] = NewTBAThr(v.dl, v.ul)
			found = true
			break
		}
	}
	if !found {
		f.connThr[conn] = &unlimtedBandwith
	}
	return nil
}

func (f *CIDRLimit) GetConnLimit(conn net.Conn) *BandThrot {
	lim := &unlimtedBandwith

	f.mtx.RLock()
	defer f.mtx.RUnlock()

	if connLim, ok := f.connThr[conn]; ok {
		lim = connLim
	}
	return lim
}

func (f *CIDRLimit) RemoveConn(conn net.Conn) {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	delete(f.connThr, conn)
}
