package core

import (
	"github.com/gin-gonic/gin"
	"github.com/lp2p/p2pvpn/api/route"
	"github.com/lp2p/p2pvpn/constant"
)

type APIService struct {
	router *gin.Engine
	addr   string
	tab    *route.Table
}

// NewDefaultAPIService create a APIService using gin.Default,
// with Logger and Recovery.
func NewDefaultAPIService(addr string) *APIService {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	tab := route.NewRouteTable()
	return NewAPIService(router, tab, addr)
}

// NewAPIService create a APIService with provider gin.Engine and route.RouteTable,
// it's convenient for testing.
func NewAPIService(router *gin.Engine, tab *route.Table, addr string) *APIService {
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
}

// Run starts api service.
func (a *APIService) Run() {
	a.RegisterHandler()
	err := a.router.Run(a.addr)
	if err != nil {
		panic(err)
	}
}
