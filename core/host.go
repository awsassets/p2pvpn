package core

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/p2p/host/relay"
	"strings"

	"github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/lp2p/p2pvpn/api/route"
	"github.com/lp2p/p2pvpn/common/utils"
	"github.com/lp2p/p2pvpn/log"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

var _host host.Host = nil

func Host() host.Host {
	return _host
}

// InitHost creates a libp2p host with a generated identity.
func InitHost(address string, port int) host.Host {
	h, err := libp2p.New(context.Background(),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", address, port)))
	if err != nil {
		log.Fatalf("%v", err)
	}
	return h
}

// NewServerHost creates a libp2p host as relay.
func NewServerHost(addr string, port int, apiPort int) host.Host {
	server := fmt.Sprintf("http://127.0.0.1:%d", apiPort)

	publicIP := utils.GetPublicIP()

	h, err := libp2p.New(context.Background(),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", addr, port)),
		libp2p.EnableRelay(circuit.OptHop),
		libp2p.Routing(route.MakeRouting(server, relay.RelayRendezvous)),
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
	_host = h
	return h
}
