package route

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	"github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/relay"
	"github.com/lp2p/p2pvpn/log"
	"github.com/lp2p/p2pvpn/server"
	ma "github.com/multiformats/go-multiaddr"
)

func init() {
	relay.BootDelay = 1 * time.Second
	relay.AdvertiseBootDelay = 100 * time.Millisecond
	gin.SetMode(gin.ReleaseMode)
}

func connect(a, b host.Host) {
	peerInfo := peer.AddrInfo{ID: a.ID(), Addrs: a.Addrs()}
	err := b.Connect(context.Background(), peerInfo)
	if err != nil {
		panic(err)
	}
}

func initApiServer(addr string) {
	router := gin.New()
	tab := server.NewRouteTable()
	api := server.NewAPIService(router, tab, addr, "test")
	go api.Run()
}

func TestRouteRelay(t *testing.T) {
	log.SetAllLoggers(logging.LevelWarn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initApiServer(":8000")
	serverUrl := "http://127.0.0.1:8000"

	h1, err := libp2p.New(ctx, libp2p.EnableRelay())
	if err != nil {
		t.Fatal(err)
	}

	_, err = libp2p.New(ctx,
		libp2p.EnableRelay(circuit.OptHop),
		libp2p.Routing(MakeRouting(serverUrl, relay.RelayRendezvous, "", "test")),
		libp2p.EnableAutoRelay(),
		libp2p.AddrsFactory(func(addresses []ma.Multiaddr) []ma.Multiaddr {
			for i, addr := range addresses {
				addrString := addr.String()
				if strings.HasPrefix(addrString, "/ip4/127.0.0.1/") {
					addrNoIP := strings.TrimPrefix(addrString, "/ip4/127.0.0.1")
					addresses[i] = ma.StringCast("/dns4/localhost" + addrNoIP)
				}
			}
			return addresses
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	h3, err := libp2p.New(ctx,
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelay(),
		libp2p.Routing(MakeRouting(serverUrl, "clients", "", "test")))
	if err != nil {
		t.Fatal(err)
	}

	h4, err := libp2p.New(ctx, libp2p.EnableRelay())
	if err != nil {
		t.Fatal(err)
	}

	// connect to AutoNAT, have it resolve to private.
	connect(h1, h3)
	time.Sleep(300 * time.Millisecond)
	privateEmitter, err := h3.EventBus().Emitter(new(event.EvtLocalReachabilityChanged))
	if err != nil {
		t.Fatal(err)
	}
	err = privateEmitter.Emit(event.EvtLocalReachabilityChanged{Reachability: network.ReachabilityPrivate})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(3000 * time.Millisecond)

	unspecificRelay, err := ma.NewMultiaddr("/p2p-circuit")
	if err != nil {
		t.Fatal(err)
	}

	// verify that we now advertise relay addrs (but not unspecific relay addrs)
	haveRelay := false
	for _, addr := range h3.Addrs() {
		if addr.Equal(unspecificRelay) {
			t.Fatal("unspecific relay addr advertised")
		}

		_, err := addr.ValueForProtocol(ma.P_CIRCUIT)
		if err == nil {
			haveRelay = true
		}
	}

	if !haveRelay {
		t.Fatal("No relay addrs advertised")
	}

	var remoteAddresses []ma.Multiaddr
	for _, addr := range h3.Addrs() {
		_, err := addr.ValueForProtocol(ma.P_CIRCUIT)
		if err == nil {
			remoteAddresses = append(remoteAddresses, addr)
		}
	}

	// verify that we can connect through the relay
	err = h4.Connect(ctx, peer.AddrInfo{ID: h3.ID(), Addrs: remoteAddresses})
	if err != nil {
		t.Fatal(err)
	}

	// verify that we have pushed relay addrs to connected peers
	haveRelay = false
	for _, addr := range h1.Peerstore().Addrs(h3.ID()) {
		if addr.Equal(unspecificRelay) {
			t.Fatal("unspecific relay addr advertised")
		}

		_, err := addr.ValueForProtocol(ma.P_CIRCUIT)
		if err == nil {
			haveRelay = true
		}
	}

	if !haveRelay {
		t.Fatal("No relay addrs pushed")
	}
}

func TestConnectByID(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initApiServer(":8001")
	serverUrl := "http://127.0.0.1:8001"

	h1, err := libp2p.New(ctx,
		libp2p.Routing(MakeRouting(serverUrl, "client", "", "test")),
	)
	if err != nil {
		t.Fatal(err)
	}

	h2, err := libp2p.New(ctx,
		libp2p.Routing(MakeRouting(serverUrl, "client", "", "test")))
	if err != nil {
		t.Fatal(err)
	}

	h2Info := peer.AddrInfo{
		ID: h2.ID(),
	}

	err = h1.Connect(ctx, h2Info)
	if err != nil {
		t.Fatal(err)
	}
}
