package server

import (
	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/lp2p/p2pvpn/constant"
)

type APIService struct {
	router   *gin.Engine
	addr     string
	tab      *Table
	serverID peer.ID
}

// NewDefaultAPIService create a APIService using gin.Default,
// with Logger and Recovery.
func NewDefaultAPIService(addr string) *APIService {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	tab := NewRouteTable()
	return NewAPIService(router, tab, addr)
}

// NewAPIService create a APIService with provider gin.Engine and route.RouteTable,
// it's convenient for testing.
func NewAPIService(router *gin.Engine, tab *Table, addr string) *APIService {
	return &APIService{
		router: router,
		addr:   addr,
		tab:    tab,
	}
}

// RegisterHandler registers api service handlers to router.
func (a *APIService) RegisterHandler() {
	a.router.GET(constant.RoutingUrl+":id", a.GetNode)
	a.router.POST(constant.RoutingUrl+":cid", a.NewNode)

	a.router.GET(constant.RoutingProviderUrl+":cid", a.GetProvider)

	a.router.GET(constant.FingerprintsUrl+":fingerprint", a.GetPeerID)

	a.router.GET(constant.ServerIDUrl, a.GetServerID)
	a.router.POST(constant.ServerIDUrl+":id", a.SetServerID)
}

// Run starts api service.
func (a *APIService) Run() {
	a.RegisterHandler()
	err := a.router.Run(a.addr)
	if err != nil {
		panic(err)
	}
}
