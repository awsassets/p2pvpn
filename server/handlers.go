package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p-core/peer"
)

// GetNode returns node information by node id.
func (a *APIService) GetNode(c *gin.Context) {
	id, err := peer.Decode(c.Param("id"))
	if err != nil {
		falseResponse(http.StatusInternalServerError, c)
		return
	}
	info, err := a.tab.Find(id)
	if err != nil {
		falseResponse(http.StatusNotFound, c)
	} else {
		c.JSON(http.StatusOK, PeerResp{
			Status:   true,
			AddrInfo: info,
		})
	}
}

// NewNode creates node entry by cid and id.
func (a *APIService) NewNode(c *gin.Context) {
	addrs := c.PostForm("addrs")
	cid := c.Param("cid")
	fingerprint := c.PostForm("fingerprint")
	id, err := peer.Decode(c.PostForm("id"))
	if err != nil {
		falseResponse(http.StatusInternalServerError, c)
		return
	}

	publicAddr := strings.Split(c.Request.RemoteAddr, ":")
	strings.Replace(addrs, "127.0.0.1", publicAddr[0], 1)

	err = a.tab.Provide(cid, id, addrs, fingerprint)
	if err != nil {
		falseResponse(http.StatusInternalServerError, c)
	} else {
		c.JSON(http.StatusOK, StatusResp{
			Status: true,
		})
	}
}

// GetProvider returns all nodes under the same provider.
func (a *APIService) GetProvider(c *gin.Context) {
	cid := c.Param("cid")
	pmap, err := a.tab.FindProvider(cid)
	if err != nil {
		falseResponse(http.StatusNotFound, c)
	} else {
		c.JSON(http.StatusOK, ProvidersResp{
			Status:    true,
			AddrInfos: pmap,
		})
	}
}

// GetPeerID returns peer id by fingerprint.
func (a *APIService) GetPeerID(c *gin.Context) {
	fingerprint := c.Param("fingerprint")
	id := a.tab.FindPeerID(fingerprint)
	c.JSON(http.StatusOK, IDResp{
		PeerID: id,
	})
}

func (a *APIService) SetServerID(c *gin.Context) {
	id, err := peer.Decode(c.Param("id"))
	if err != nil {
		falseResponse(http.StatusInternalServerError, c)
		return
	}
	a.serverID = id
	c.JSON(http.StatusOK, StatusResp{
		Status: true,
	})
}

func (a *APIService) GetServerID(c *gin.Context) {
	c.JSON(http.StatusOK, IDResp{
		PeerID: a.serverID,
	})
}

// falseResponse returns false status json response.
func falseResponse(status int, c *gin.Context) {
	c.JSON(status, StatusResp{
		Status: false,
	})
}
