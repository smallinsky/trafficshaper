package trafficshaper

import (
	"net"
	"time"

	"github.com/smallinsky/trafficshaper/bandwidth"
)

const (
	maxChunkSize = 1024 * 64 // 64 KB
)

var _ net.Conn = (*connProxy)(nil)

type connProxy struct {
	c  net.Conn
	bs bandwidth.Shaper
}

func (c *connProxy) Read(b []byte) (int, error) {
	chunkSize := c.bs.GetDLCapaity(c)
	if chunkSize > maxChunkSize {
		chunkSize = maxChunkSize
	}

	i := 0
	sum := 0
	for ; i < len(b)/int(chunkSize); i++ {
		n, err := c.c.Read(b[i*int(chunkSize) : (i+1)*int(chunkSize)])
		if err != nil {
			return sum, err
		}
		c.bs.LimitDL(c, chunkSize)
		sum += n
	}

	if sum != len(b) {
		left := len(b) - sum
		c.bs.LimitDL(c, int64(left))
		n, err := c.c.Read(b[sum:len(b)])
		if err != nil {
			return sum, err
		}
		sum += n
	}
	return sum, nil
}

func (c *connProxy) Write(b []byte) (int, error) {
	chunkSize := c.bs.GetULCapaity(c)
	if chunkSize > maxChunkSize {
		chunkSize = maxChunkSize
	}

	i := 0
	sum := 0

	for ; i < len(b)/int(chunkSize); i++ {
		n, err := c.c.Write(b[i*int(chunkSize) : (i+1)*int(chunkSize)])
		if err != nil {
			return sum, err
		}
		c.bs.LimitUL(c, chunkSize)
		sum += n
	}

	if sum != len(b) {
		left := len(b) - sum
		c.bs.LimitUL(c, int64(left))
		n, err := c.c.Write(b[sum:len(b)])
		if err != nil {
			return sum, err
		}
		sum += n
	}
	return sum, nil
}

func (c *connProxy) Close() error {
	c.bs.RemoveConn(c)
	return c.c.Close()
}

func (c *connProxy) LocalAddr() net.Addr {
	return c.c.LocalAddr()
}

func (c *connProxy) RemoteAddr() net.Addr {
	return c.c.RemoteAddr()
}

func (c *connProxy) SetDeadline(t time.Time) error {
	return c.c.SetDeadline(t)
}

func (c *connProxy) SetReadDeadline(t time.Time) error {
	return c.c.SetReadDeadline(t)
}

func (c *connProxy) SetWriteDeadline(t time.Time) error {
	return c.c.SetWriteDeadline(t)
}
