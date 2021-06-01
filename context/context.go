package context

import "net"

type ConnContext struct {
	Addr net.Addr
	Conn net.Conn
}
