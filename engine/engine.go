package engine

import (
	"errors"
	"net"

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
	SocksAddr string
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

				tunnel.Add(context.ConnContext{Addr: &tcpAddr{target.String()}, Conn: conn})
			}()
		}
	}()

	return nil
}
