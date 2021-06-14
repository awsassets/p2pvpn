package engine

type tcpAddr struct {
	addr string
}

func (*tcpAddr) Network() string {
	return "tcp"
}

func (a *tcpAddr) String() string {
	return a.addr
}
