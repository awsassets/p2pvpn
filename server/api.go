package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/lp2p/p2pvpn/common/utils"
	"github.com/lp2p/p2pvpn/constant"
)

type APIService struct {
	router   *gin.Engine
	addr     string
	tab      *Table
	serverID peer.ID
	secret   string
}

// NewDefaultAPIService create a APIService using gin.Default,
// with Logger and Recovery.
func NewDefaultAPIService(addr string, secret string) *APIService {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	tab := NewRouteTable()
	return NewAPIService(router, tab, addr, secret)
}

// NewAPIService create a APIService with provider gin.Engine and route.RouteTable,
// it's convenient for testing.
func NewAPIService(router *gin.Engine, tab *Table, addr, secret string) *APIService {
	return &APIService{
		router: router,
		addr:   addr,
		tab:    tab,
		secret: secret,
	}
}

// RegisterHandler registers api service handlers to router.
func (a *APIService) RegisterHandler() {
	a.router.GET(constant.RoutingUrl+":id", a.GetPeer)
	a.router.POST(constant.RoutingUrl+":cid", a.NewPeer)

	a.router.GET(constant.RoutingProviderUrl+":cid", a.GetProvider)

	a.router.GET(constant.FingerprintsUrl+":fingerprint", a.GetPeerID)
	a.router.DELETE(constant.FingerprintsUrl+":fingerprint", a.DeletePeer)

	a.router.GET(constant.ServerIDUrl, a.GetServerID)
	a.router.POST(constant.ServerIDUrl+":id", a.SetServerID)
}

// Run starts api service.
func (a *APIService) Run() {
	a.router.Use(a.Auth())
	a.RegisterHandler()
	err := a.router.Run(a.addr)
	if err != nil {
		panic(err)
	}
}

// Auth is a gin middleware to auth request.
func (a *APIService) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientSecret := c.GetHeader("auth")
		secret := utils.Md5(a.secret)
		if clientSecret == secret {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, StatusResp{
				Status: false,
			})
		}
	}
}
