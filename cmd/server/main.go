package main

import (
	"flag"
	"fmt"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/p2p/host/relay"
	"github.com/lp2p/p2pvpn/core"
	"github.com/lp2p/p2pvpn/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	// Reset delay, so we can advertise ourselves immediately
	relay.BootDelay = 1 * time.Second
	relay.AdvertiseBootDelay = 100 * time.Millisecond
}

func main() {
	log.SetAllLoggers(logging.LevelWarn)

	addr := flag.String("i", "0.0.0.0", "server ip address")
	p2pPort := flag.Int("l", 12001, "libp2p server port")
	apiPort := flag.Int("a", 8000, "api service port")
	flag.Parse()

	api := core.NewDefaultAPIService(fmt.Sprintf(":%d", *apiPort))
	go api.Run()
	go core.NewServerHost(*addr, *p2pPort)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
