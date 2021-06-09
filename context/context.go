package context

import (
	"net"
	"time"
)

type Addr = net.Addr

type Conn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

type ConnContext struct {
	Addr Addr
	Conn Conn
}
