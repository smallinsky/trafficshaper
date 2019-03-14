package bandwidth

import (
	"net"
	"testing"
	"time"
)

func TestCIRDLimit(t *testing.T) {
	checkConnBand := func(b *BandThrot, dl, ul int64) {
		if got, expected := b.DL(), dl; got != expected {
			t.Fatalf("Got %v, expected %v", got, expected)
		}
		if got, expected := b.UL(), ul; got != expected {
			t.Fatalf("Got %v, expected %v", got, expected)
		}
	}

	t.Run("AddConnAfterLimitIsAddied", func(t *testing.T) {
		sut := NewCIDRLimit()
		conn := &connMock{
			remoteAddr: net.TCPAddr{
				IP: net.ParseIP("127.0.0.1"),
			},
		}

		if err := sut.AddLimit("127.0.0.0/24", 1024, 1024); err != nil {
			t.Fatalf("Faild to add conneciton limit: %v", err)
		}
		if err := sut.AddConn(conn); err != nil {
			t.Fatal("failed to add connection: ", err)
		}

		checkConnBand(sut.GetConnLimit(conn), 1024, 1024)

	})
	t.Run("AddConnBeforeLimitIsAppied", func(t *testing.T) {
		sut := NewCIDRLimit()
		conn := &connMock{
			remoteAddr: net.TCPAddr{
				IP: net.ParseIP("127.0.0.1"),
			},
		}

		if err := sut.AddConn(conn); err != nil {
			t.Fatal("failed to add connection: ", err)
		}
		if err := sut.AddLimit("127.0.0.0/24", 1024, 1024); err != nil {
			t.Fatalf("Faild to add conneciton limit: %v", err)
		}

		checkConnBand(sut.GetConnLimit(conn), 1024, 1024)
	})

	t.Run("AddSecondLimitForCIDRPool", func(t *testing.T) {
		sut := NewCIDRLimit()
		conn := &connMock{
			remoteAddr: net.TCPAddr{
				IP: net.ParseIP("127.0.0.1"),
			},
		}

		if err := sut.AddConn(conn); err != nil {
			t.Fatal("failed to add connection: ", err)
		}

		if err := sut.AddLimit("127.0.0.0/24", 1024, 1024); err != nil {
			t.Fatalf("Faild to add conneciton limit: %v", err)
		}

		checkConnBand(sut.GetConnLimit(conn), 1024, 1024)

		if err := sut.AddLimit("127.0.0.0/24", 2048, 2048); err != nil {
			t.Fatalf("Faild to add conneciton limit: %v", err)
		}

		checkConnBand(sut.GetConnLimit(conn), 2048, 2048)
	})

	t.Run("MultileConnections", func(t *testing.T) {
		sut := NewCIDRLimit()
		firstConn := &connMock{
			remoteAddr: net.TCPAddr{
				IP: net.ParseIP("127.0.0.1"),
			},
		}
		secConn := &connMock{
			remoteAddr: net.TCPAddr{
				IP: net.ParseIP("192.0.0.1"),
			},
		}
		thirdConn := &connMock{
			remoteAddr: net.TCPAddr{
				IP: net.ParseIP("127.0.0.2"),
			},
		}

		if err := sut.AddConn(firstConn); err != nil {
			t.Fatal("failed to add connection: ", err)
		}

		if err := sut.AddConn(secConn); err != nil {
			t.Fatal("failed to add connection: ", err)
		}

		if err := sut.AddConn(thirdConn); err != nil {
			t.Fatal("failed to add connection: ", err)
		}

		if err := sut.AddLimit("127.0.0.0/24", 1024, 1024); err != nil {
			t.Fatalf("Faild to add conneciton limit: %v", err)
		}

		if err := sut.AddLimit("192.0.0.0/24", 9092, 9092); err != nil {
			t.Fatalf("Faild to add conneciton limit: %v", err)
		}

		checkConnBand(sut.GetConnLimit(firstConn), 1024, 1024)
		checkConnBand(sut.GetConnLimit(secConn), 9092, 9092)
		checkConnBand(sut.GetConnLimit(thirdConn), 1024, 1024)
	})

	t.Run("NotAddedConn", func(t *testing.T) {
		sut := NewCIDRLimit()
		firstConn := &connMock{
			remoteAddr: net.TCPAddr{
				IP: net.ParseIP("127.0.0.1"),
			},
		}

		if err := sut.AddLimit("127.0.0.0/24", 1024, 1024); err != nil {
			t.Fatalf("Faild to add conneciton limit: %v", err)
		}

		checkConnBand(sut.GetConnLimit(firstConn), unlimtedBandwith.DL(), unlimtedBandwith.UL())
	})

	t.Run("ConnWithoutLimit", func(t *testing.T) {
		sut := NewCIDRLimit()
		firstConn := &connMock{
			remoteAddr: net.TCPAddr{
				IP: net.ParseIP("127.0.0.1"),
			},
		}

		checkConnBand(sut.GetConnLimit(firstConn), unlimtedBandwith.DL(), unlimtedBandwith.UL())
	})

}

type connMock struct {
	localAddr  net.TCPAddr
	remoteAddr net.TCPAddr
}

func (m *connMock) Read(b []byte) (int, error) {
	return 0, nil
}

func (m *connMock) Write(b []byte) (int, error) {
	return 0, nil
}

func (m *connMock) Close() error {
	return nil
}

func (m *connMock) LocalAddr() net.Addr {
	return &m.localAddr
}

func (m *connMock) RemoteAddr() net.Addr {
	return &m.remoteAddr
}

func (m *connMock) SetDeadline(t time.Time) error {
	return nil
}

func (m *connMock) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *connMock) SetWriteDeadline(t time.Time) error {
	return nil
}

type addrMock struct {
	network string
	ip      string
}

func (m *addrMock) Network() string {
	return m.network
}

func (m *addrMock) String() string {
	return m.ip
}
