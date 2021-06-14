package core

import (
	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/lp2p/p2pvpn/api/route"
	"github.com/lp2p/p2pvpn/constant"
	"net/http"
	"strings"
)

type APIService struct {
	router *gin.Engine
	addr   string
	tab    *route.RouteTable
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
func NewAPIService(router *gin.Engine, tab *route.RouteTable, addr string) *APIService {
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
}

// GetNode returns node information by node id.
func (a *APIService) GetNode(c *gin.Context) {
	id, err := peer.Decode(c.Param("id"))
	if err != nil {
		falseResponse(c)
		return
	}
	info, err := a.tab.Find(id)
	if err != nil {
		c.JSON(http.StatusNotFound, route.StatusResp{
			Status: false,
		})
	} else {
		c.JSON(http.StatusOK, route.PeerResp{
			Status:   true,
			AddrInfo: info,
		})
	}
}

// NewNode creates node entry by cid and id.
func (a *APIService) NewNode(c *gin.Context) {
	addrs := c.PostForm("addrs")
	cid := c.Param("cid")
	id, err := peer.Decode(c.PostForm("id"))
	if err != nil {
		falseResponse(c)
		return
	}

	publicAddr := strings.Split(c.Request.RemoteAddr, ":")
	strings.Replace(addrs, "127.0.0.1", publicAddr[0], 1)

	err = a.tab.Provide(cid, id, addrs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, route.StatusResp{
			Status: false,
		})
	} else {
		c.JSON(http.StatusOK, route.StatusResp{
			Status: true,
		})
	}
}

// GetProvider returns all nodes under the same provider.
func (a *APIService) GetProvider(c *gin.Context) {
	cid := c.Param("cid")
	pmap, err := a.tab.FindProvider(cid)
	if err != nil {
		c.JSON(http.StatusNotFound, route.StatusResp{
			Status: false,
		})
	} else {
		c.JSON(http.StatusOK, route.ProvidersResp{
			Status:    true,
			AddrInfos: pmap,
		})
	}
}

// Run starts api service.
func (a *APIService) Run() {
	a.RegisterHandler()
	err := a.router.Run(a.addr)
	if err != nil {
		panic(err)
	}
}

func falseResponse(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, route.StatusResp{
		Status: false,
	})
}
