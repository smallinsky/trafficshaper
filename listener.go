package trafficshaper

import (
	"fmt"
	"net"

	"github.com/smallinsky/trafficshaper/bandwidth"
)

func NewListener(network, address string, opts ...Opt) (*Listener, error) {
	var options options
	for _, o := range opts {
		o(&options)
	}

	bs := bandwidth.NewShaper()
	if err := applyOptions(bs, options); err != nil {
		return nil, err
	}

	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	return &Listener{
		l:  l,
		bs: bs,
	}, nil
}

func FromListener(l net.Listener, opts ...Opt) (*Listener, error) {
	var options options
	for _, o := range opts {
		o(&options)
	}

	bs := bandwidth.NewShaper()
	if err := applyOptions(bs, options); err != nil {
		return nil, err
	}

	return &Listener{
		l:  l,
		bs: bs,
	}, nil
}

func applyOptions(bs bandwidth.Shaper, ops options) error {
	if ops.serverLimit != nil {
		if err := bs.AddServerLimit(ops.serverLimit.DL, ops.serverLimit.UL); err != nil {
			return fmt.Errorf("failed to add server limit: %v", err)
		}
	}
	for _, c := range ops.cidrs {
		if err := bs.AddCIDRLimit(c.cidr, c.limit.DL, c.limit.UL); err != nil {
			return fmt.Errorf("failed to add crid limit: %v", err)
		}
	}
	return nil
}

var _ net.Listener = (*Listener)(nil)

type Listener struct {
	l  net.Listener
	bs bandwidth.Shaper
}

func (c *Listener) AddServerLimit(dl, ul int64) {
	c.bs.AddServerLimit(dl, ul)
}

func (s *Listener) AddCIDRLimit(cidr string, dl, ul int64) error {
	return s.bs.AddCIDRLimit(cidr, dl, ul)
}

func (n *Listener) Accept() (net.Conn, error) {
	conn, err := n.l.Accept()
	if err != nil {
		return nil, err
	}
	cw := &connProxy{
		c:  conn,
		bs: n.bs,
	}

	n.bs.AddConn(cw)
	return cw, err

}

func (n *Listener) Close() error {
	return n.l.Close()
}

func (n *Listener) Addr() net.Addr {
	return n.l.Addr()
}
