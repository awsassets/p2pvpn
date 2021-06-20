package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/p2p/host/relay"
	"github.com/lp2p/p2pvpn/constant"
	"github.com/lp2p/p2pvpn/core"
	"github.com/lp2p/p2pvpn/log"
	"github.com/lp2p/p2pvpn/server"
)

func init() {
	// Reset delay, so we can advertise ourselves immediately
	relay.BootDelay = 1 * time.Second
	relay.AdvertiseBootDelay = 100 * time.Millisecond
}

func main() {
	log.SetAllLoggers(logging.LevelWarn)

	apiPort := flag.Int("api-port", 8000, "api service port")
	secret := flag.String("secret", constant.DefaultSecret, "api auth secret")
	flag.Parse()

	api := server.NewDefaultAPIService(fmt.Sprintf(":%d", *apiPort), *secret)
	go api.Run()
	go core.NewServerHost(*apiPort, *secret)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
