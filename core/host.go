package core

import (
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
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
		log.Fatalln(err)
	}
	_host = h
	return h
}
