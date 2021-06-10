package core

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/lp2p/p2pvpn/log"
	"github.com/lp2p/p2pvpn/route"
	"github.com/lp2p/p2pvpn/utils"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"strings"
)

// NewHost creates a libp2p host with a generated identity.
func NewHost(address string, port int) host.Host {
	h, err := libp2p.New(context.Background(),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", address, port)))
	if err != nil {
		log.Fatalf("%v", err)
	}
	return h
}

// NewServerHost creates a libp2p host as relay.
func NewServerHost(addr string, port int) host.Host {
	var router routing.PeerRouting
	makeRouting := func(h host.Host) (routing.PeerRouting, error) {
		router = route.NewRoute(h, "http://127.0.0.1:8000")
		return router, nil
	}

	publicIP := utils.GetPublicIP()

	h, err := libp2p.New(context.Background(),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", addr, port)),
		libp2p.EnableRelay(circuit.OptHop),
		libp2p.Routing(makeRouting),
		libp2p.EnableAutoRelay(),
		libp2p.AddrsFactory(func(addrs []ma.Multiaddr) []ma.Multiaddr {
			hasPublicIP := false
			var address ma.Multiaddr
			index := 0
			for i, addr := range addrs {
				if manet.IsPublicAddr(addr) {
					hasPublicIP = true
				}
				addrString := addr.String()
				if strings.HasPrefix(addrString, "/ip4/127.0.0.1/") {
					index = i
					address = addr
				}
			}
			if !hasPublicIP && address != nil {
				addrString := address.String()
				addrNoIP := strings.TrimPrefix(addrString, "/ip4/127.0.0.1/")
				addrs[index] = ma.StringCast(fmt.Sprintf("/ip4/%s/%s", publicIP, addrNoIP))
			}
			return addrs
		}),
	)
	if err != nil {
		log.Errorf("%v", err)
	}
	return h
}
