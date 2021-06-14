package engine

import (
	gocontext "context"
	"errors"
	"net"
	"strconv"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/lp2p/p2pvpn/constant"
	"github.com/lp2p/p2pvpn/context"
	"github.com/lp2p/p2pvpn/core"
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
	SocksAddr  string
	ServerAddr string
}

type engine struct {
	*Key
}

func (e *engine) start() error {
	if e.Key == nil {
		return errors.New("empty key")
	}

	for _, f := range []func() error{
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

				/*
					HOOK DEST PEER
				*/
				var dest peer.ID

				stream, err := core.Host().NewStream(gocontext.Background(), dest, constant.Protocol)
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
	host, port, err := net.SplitHostPort(e.ServerAddr)
	if err != nil {
		return err
	}
	portInt, _ := strconv.Atoi(port)

	h := core.InitHost(host, portInt)

	// We let our host know that it needs to handle streams tagged with the
	// protocol id that we have defined, and then handle them to
	// our own streamHandling function.
	h.SetStreamHandler(constant.Protocol, func(stream network.Stream) {
		buf := make([]byte, socks5.MaxAddrLen)

		addr, err := socks5.ReadAddr(stream, buf)
		if err != nil {
			log.Warnf("Read address failed: %v", err)
			return
		}

		tunnel.Add(context.ConnContext{Addr: &tcpAddr{addr.String()}, Conn: stream})
	})

	log.Infof("Peer host is listening at:")
	for _, a := range h.Addrs() {
		log.Infof("%s/%s\n", a, peer.Encode(h.ID()))
	}

	return nil
}
