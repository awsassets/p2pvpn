package main

import (
	"flag"
	"fmt"

	"github.com/lp2p/p2pvpn/core"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/lp2p/p2pvpn/log"
)

func main() {
	// Parse some flags
	destPeer := flag.String("d", "", "destination peer address")
	port := flag.Int("p", 9900, "proxy port")
	p2pport := flag.Int("l", 12000, "libp2p listen port")
	flag.Parse()

	// If we have a destination peer we will start a local server
	if *destPeer != "" {
		// We use p2pport+1 in order to not collide if the user
		// is running the remote peer locally on that port
		host := core.NewHost("0.0.0.0", *p2pport+1)
		// Make sure our host knows how to reach destPeer
		destPeerID := core.AddAddrToPeerstore(host, *destPeer)
		proxyAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", *port))
		if err != nil {
			log.Errorf("%v", err)
		}
		// Create the proxy service and start the http server
		proxy := core.NewProxyService(host, proxyAddr, destPeerID)
		proxy.Serve() // serve hangs forever
	} else {
		host := core.NewHost("0.0.0.0", *p2pport)
		// In this case we only need to make sure our host
		// knows how to handle incoming proxied requests from
		// another peer.
		_ = core.NewProxyService(host, nil, "")
		<-make(chan struct{}) // hang forever
	}

	//sigCh := make(chan os.Signal, 1)
	//signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	//<-sigCh
}
