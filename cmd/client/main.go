package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/lp2p/p2pvpn/engine"
	"github.com/lp2p/p2pvpn/log"
)

var key = new(engine.Key)

func init() {
	flag.StringVar(&key.SocksAddr, "socks-addr", "", "socks addr to bind")
	flag.Parse()
}

func main() {
	engine.Insert(key)

	checkErr := func(msg string, f func() error) {
		if err := f(); err != nil {
			log.Fatalf("Failed to %s: %v", msg, err)
		}
	}

	checkErr("start engine", engine.Start)
	defer checkErr("stop engine", engine.Stop)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
