package trafficshaper

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func startServer(t *testing.T, buffSize int, clientBandwith, serverBadwith int64) func() {
	opts := []Opt{
		WithServerBandwith(Limit{UL: serverBadwith, DL: serverBadwith}),
		WithCIDRBandiwth("127.0.0.1/26", Limit{UL: clientBandwith, DL: clientBandwith}),
	}

	ln, err := NewListener("tcp", "127.0.0.1:8080", opts...)
	if err != nil {
		panic(err)
	}

	stopC := make(chan struct{})
	stopFn := func() {
		ln.Close()
		close(stopC)
	}

	buf := make([]byte, buffSize)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				for {
					select {
					case <-stopC:
						conn.Close()
						return
					default:
						_, err := conn.Write(buf)
						if err != nil {
							return
						}
					}
				}
			}()
		}
	}()
	return stopFn
}

func TestClient(t *testing.T) {
	tests := []struct {
		name                string
		clientReadBuffSize  int
		serverWriteBuffSize int
		numOfClients        int
		waitTime            time.Duration
		marginError         float64
		clientBandwith      int64
		serverBadwith       int64
	}{
		{
			name:                "HundredClientsWith10KBLimit",
			clientReadBuffSize:  1024,
			serverWriteBuffSize: 1024,
			numOfClients:        100,
			waitTime:            time.Second * 2,
			clientBandwith:      10 * KBps,
			serverBadwith:       1000 * KBps,
			marginError:         0.05,
		},
		{
			name:                "ThirtyClientsWith10MBLimit",
			clientReadBuffSize:  5024,
			serverWriteBuffSize: 5024,
			numOfClients:        30,
			waitTime:            time.Second * 2,
			clientBandwith:      10 * MBps,
			serverBadwith:       300 * MBps,
			marginError:         0.05,
		},
	}

	for _, tc := range tests {
		stopSrv := startServer(t, tc.serverWriteBuffSize, tc.clientBandwith, tc.serverBadwith)
		stopSrv()

		t.Run(tc.name, func(t *testing.T) {
			startCtx, start := context.WithCancel(context.Background())

			sumMap := make(map[int]int)
			var mtx sync.RWMutex

			stopSrv := startServer(t, tc.serverWriteBuffSize, tc.clientBandwith, tc.serverBadwith)

			for i := 0; i < tc.numOfClients; i++ {
				go func(startCtx context.Context, i int) {
					conn, err := net.Dial("tcp", "localhost:8080")
					if err != nil {
						t.Fatalf("failed to connect to server: %v", err)
					}

					buff := make([]byte, tc.clientReadBuffSize)
					sum := 0

					<-startCtx.Done()
					for {
						n, err := conn.Read(buff)
						if err != nil {
							return
						}
						sum += n
						mtx.Lock()
						sumMap[i] = sum
						mtx.Unlock()
					}
				}(startCtx, i)
			}

			start()
			time.Sleep(tc.waitTime)

			shouldRead := int(tc.waitTime.Seconds() * float64(tc.clientBandwith))
			margin := int(float64(shouldRead) * tc.marginError)

			t.Logf("Total margin error: %v  average client margin error: %v", margin, margin/tc.numOfClients)
			mtx.Lock()
			for _, sum := range sumMap {
				t.Logf("client read %v bytes bandwith %v KB/s", sum, float64(sum)/(tc.waitTime.Seconds())/1024)
				if sum < shouldRead-margin || sum > shouldRead+margin {
					t.Fatalf("Expected %v bytes, Got %v bytes", shouldRead, sum)
				}
				t.Logf("Should read %v, actual %v, actual/expected %v", shouldRead, sum, float64(sum)/float64(shouldRead))
			}
			mtx.Unlock()
			stopSrv()
		})
	}
}
