package engine

import (
	gocontext "context"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-core/event"
	"github.com/lp2p/p2pvpn/common/utils"
	"net"
	"net/url"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/lp2p/p2pvpn/api/route"
	"github.com/lp2p/p2pvpn/common/pool"
	"github.com/lp2p/p2pvpn/constant"
	"github.com/lp2p/p2pvpn/context"
	"github.com/lp2p/p2pvpn/log"
	"github.com/lp2p/p2pvpn/transport/socks5"
	"github.com/lp2p/p2pvpn/tunnel"
)

var _engine = &engine{}

// Start starts the default engine up.
func Start() error {
	return _engine.start()
}

// Stop shuts the default engine down.
func Stop() error {
	return _engine.stop()
}

// Insert loads *Key to the default engine.
func Insert(k *Key) {
	_engine.insert(k)
}

type Key struct {
	SocksAddr   string
	ServerUrl   string
	Fingerprint string

	secret string
}

type engine struct {
	*Key

	host host.Host
}

func (e *engine) start() error {
	if e.Key == nil {
		return errors.New("empty key")
	}

	for _, f := range []func() error{
		e.initServerUrl,
		e.initHost,
		e.initAutoNAT,
		e.initSocks,
		e.initP2PHost,
	} {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func (e *engine) stop() error {
	return nil
}

func (e *engine) insert(k *Key) {
	e.Key = k
}

// initServerUrl gets secret from server url.
func (e *engine) initServerUrl() error {
	serverUrl, err := url.Parse(e.ServerUrl)
	if err != nil {
		return err
	}

	e.secret = serverUrl.User.Username()
	e.ServerUrl = fmt.Sprintf("%s://%s", serverUrl.Scheme, serverUrl.Host)

	return nil
}

// initHost creates a libp2p host with a generated identity.
func (e *engine) initHost() error {
	h, err := libp2p.New(gocontext.Background(),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelay(),
		libp2p.Routing(route.MakeRouting(e.ServerUrl, constant.PeerRendezvous, e.Fingerprint, e.secret)),
	)
	if err != nil {
		return err
	}

	e.host = h
	return nil
}

// initAutoNAT connect to server nat service, figure out our nat type.
func (e *engine) initAutoNAT() error {
	serverID, err := route.Router().GetServerID()
	if err != nil {
		return err
	}

	serverInfo := peer.AddrInfo{
		ID: serverID,
	}

	err = e.host.Connect(gocontext.Background(), serverInfo)
	if err != nil {
		return err
	}

	go e.listenNATChange()

	return nil
}

func (e *engine) initSocks() error {
	l, err := net.Listen("tcp", e.SocksAddr)
	if err != nil {
		return err
	}

	log.Infof("SOCKS proxy listening at: %s", e.SocksAddr)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Debugf("SOCKS accept error: %v", err)
				continue
			}

			go func() {
				target, command, err := socks5.ServerHandshake(conn)
				if err != nil || command != socks5.CmdConnect {
					_ = conn.Close()
					return
				}

				if c, ok := conn.(*net.TCPConn); ok {
					_ = c.SetKeepAlive(true)
				}

				stream, err := e.newStream(target)
				// If an error happens, we write an error for response.
				if err != nil {
					log.Warnf("Starting new stream failed: %v", err)
					return
				}

				log.Infof("New stream connection: %s <--> %s", conn.RemoteAddr(), stream.ID())

				defer conn.Close()
				defer stream.Close()

				stream.Write(target)
				tunnel.Relay(conn, stream)
			}()
		}
	}()

	return nil
}

func (e *engine) initP2PHost() error {
	// We let our host know that it needs to handle streams tagged with the
	// protocol id that we have defined, and then handle them to
	// our own streamHandling function.
	e.host.SetStreamHandler(constant.Protocol, func(stream network.Stream) {
		buf := pool.Get(socks5.MaxAddrLen)
		defer pool.Put(buf)

		addr, err := socks5.ReadAddr(stream, buf)
		if err != nil {
			log.Warnf("Read address failed: %v", err)
			return
		}

		addrHost, addrPort := addr.ToHostPort()
		if addrHost == e.Fingerprint {
			addrHost = net.JoinHostPort("127.0.0.1", addrPort)
		}

		tunnel.Add(context.ConnContext{Addr: &tcpAddr{addrHost}, Conn: stream})
	})

	log.Infof("Peer host is listening at:")
	for _, a := range e.host.Addrs() {
		log.Infof("%s/%s\n", a, peer.Encode(e.host.ID()))
	}

	return nil
}

// newStream creates a stream between e.host and target peer.
func (e *engine) newStream(target socks5.Addr) (network.Stream, error) {
	targetStr, _ := target.ToHostPort()
	peerID, err := route.Router().FindPeerID(targetStr)
	if err != nil {
		return nil, err
	}
	targetInfo := peer.AddrInfo{
		ID: peerID,
	}

	err = e.host.Connect(gocontext.Background(), targetInfo)
	if err != nil {
		return nil, err
	}

	stream, err := e.host.NewStream(gocontext.Background(), peerID, constant.Protocol)
	if err != nil {
		return nil, err
	}

	return stream, nil
}

// listenNATChange subscribes nat change event. When change to private, libp2p will auto
// select a node as relay, and update address, then we advertise our new addres to server.
func (e *engine) listenNATChange() {
	subscriber, err := e.host.EventBus().Subscribe(&event.EvtLocalReachabilityChanged{})
	if err != nil {
		log.Errorf("%v", err)
	}

	select {
	case ev := <-subscriber.Out():
		tureEv, ok := ev.(event.EvtLocalReachabilityChanged)
		if ok && tureEv.Reachability == network.ReachabilityPrivate {
			// Waiting for select relay.
			time.Sleep(3000 * time.Millisecond)
			cid := utils.StrToCid(constant.PeerRendezvous)
			err := route.Router().Provide(gocontext.Background(), cid, true)
			if err != nil {
				log.Errorf("%v", err)
			}
			log.Infof("Advertise relay address success")
		}
	}
}
