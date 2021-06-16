package tunnel

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/lp2p/p2pvpn/common/pool"
	"github.com/lp2p/p2pvpn/context"
	"github.com/lp2p/p2pvpn/log"
)

const (
	tcpWaitTimeout = 5 * time.Second
)

func handleTCPConn(cc context.ConnContext) {
	conn := cc.Conn
	defer conn.Close()

	c, err := net.Dial(cc.Addr.Network(), cc.Addr.String())
	if err != nil {
		log.Errorf("TUNNEL: dial %s failed: %v", cc.Addr.String(), err)
		return
	}
	defer c.Close()

	relay(conn, c)
}

// relay copies between left and right bidirectionally.
func relay(left, right context.Conn) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		_ = copyBuffer(right, left) /* ignore error */
		right.SetReadDeadline(time.Now().Add(tcpWaitTimeout))
	}()

	go func() {
		defer wg.Done()
		_ = copyBuffer(left, right) /* ignore error */
		left.SetReadDeadline(time.Now().Add(tcpWaitTimeout))
	}()

	wg.Wait()
}

func copyBuffer(dst io.Writer, src io.Reader) error {
	buf := pool.Get(pool.RelayBufferSize)
	defer pool.Put(buf)

	_, err := io.CopyBuffer(dst, src, buf)
	return err
}
