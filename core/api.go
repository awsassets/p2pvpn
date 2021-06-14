package core

import (
	"github.com/gin-gonic/gin"
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
	// Get node information
	router.GET(constant.RoutingUrl+":id", func(c *gin.Context) {
		id := c.Param("id")
		info, err := tab.Find(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status": false,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status":    true,
				"addr_info": info,
			})
		}
	})

	// Create a new node
	router.POST(constant.RoutingUrl+":cid", func(c *gin.Context) {
		cid := c.Param("cid")
		id := c.PostForm("id")
		addrs := c.PostForm("addrs")

		publicAddr := strings.Split(c.Request.RemoteAddr, ":")
		strings.Replace(addrs, "127.0.0.1", publicAddr[0], 1)

		// fmt.Println(cid, id, addrs)
		err := tab.Provide(cid, id, addrs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": false,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status": true,
			})
		}
	})

	// Get node by provider
	router.GET(constant.RoutingProviderUrl+":cid", func(c *gin.Context) {
		cid := c.Param("cid")
		pmap, err := tab.FindProvider(cid)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"status": false,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status":     true,
				"addr_infos": pmap,
			})
		}
	})

	return &APIService{
		router: router,
		addr:   addr,
		tab:    tab,
	}
}

func (a *APIService) Run() {
	err := a.router.Run(a.addr)
	if err != nil {
		panic(err)
	}
}
