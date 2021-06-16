package engine

import (
	gocontext "context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/lp2p/p2pvpn/api/route"
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
	ServerAddr  string
	Fingerprint string
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
		e.initHost,
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

// initHost creates a libp2p host with a generated identity.
func (e *engine) initHost() error {
	server := "http://" + e.ServerAddr
	h, err := libp2p.New(gocontext.Background(),
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelay(),
		libp2p.Routing(route.MakeRouting(server, constant.PeerRendezvous, e.Key.Fingerprint)),
	)

	if err != nil {
		return err
	}

	e.host = h
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
		buf := make([]byte, socks5.MaxAddrLen)

		addr, err := socks5.ReadAddr(stream, buf)
		if err != nil {
			log.Warnf("Read address failed: %v", err)
			return
		}

		addrStr := addr.String()
		addrHost, addrPort, _ := net.SplitHostPort(addrStr)
		if addrHost == e.Fingerprint {
			addrStr = fmt.Sprintf("127.0.0.1:%s", addrPort)
		}

		tunnel.Add(context.ConnContext{Addr: &tcpAddr{addrStr}, Conn: stream})
	})

	log.Infof("Peer host is listening at:")
	for _, a := range e.host.Addrs() {
		log.Infof("%s/%s\n", a, peer.Encode(e.host.ID()))
	}

	return nil
}

// newStream creates a stream between e.host and target peer.
func (e *engine) newStream(target socks5.Addr) (network.Stream, error) {
	targetStr, _, _ := net.SplitHostPort(target.String())
	resp, err := http.Get("http://" + e.ServerAddr + constant.FingerprintsUrl + targetStr)
	if err != nil {
		return nil, err
	}

	res, err := io.ReadAll(resp.Body)
	var respPtr route.IDResp
	err = json.Unmarshal(res, &respPtr)
	if err != nil {
		return nil, err
	}
	targetInfo := peer.AddrInfo{
		ID: respPtr.PeerID,
	}

	err = e.host.Connect(gocontext.Background(), targetInfo)
	if err != nil {
		return nil, err
	}

	stream, err := e.host.NewStream(gocontext.Background(), respPtr.PeerID, constant.Protocol)
	if err != nil {
		return nil, err
	}

	return stream, nil
}
