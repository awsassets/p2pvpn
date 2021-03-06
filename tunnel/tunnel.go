package tunnel

import (
	"github.com/lp2p/p2pvpn/context"
)

var (
	tcpQueue = make(chan context.ConnContext)
)

func init() {
	go process()
}

// Add request to queue
func Add(ctx context.ConnContext) {
	tcpQueue <- ctx
}

// Relay exports internal relay function.
var Relay = relay

func process() {
	for c := range tcpQueue {
		go handleTCPConn(c)
	}
}
