package server

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p-core/peer"
)

// GetPeer returns peer information by peer id.
func (a *APIService) GetPeer(c *gin.Context) {
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

// NewPeer creates peer entry by cid and peer id.
func (a *APIService) NewPeer(c *gin.Context) {
	addrs := c.PostForm("addrs")
	cid := c.Param("cid")
	fingerprint := c.PostForm("fingerprint")
	id, err := peer.Decode(c.PostForm("id"))
	if err != nil {
		falseResponse(http.StatusInternalServerError, c)
		return
	}

	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		falseResponse(http.StatusInternalServerError, c)
		return
	}
	addrs = strings.Replace(addrs, "127.0.0.1", host, 1)

	err = a.tab.Provide(cid, id, addrs, fingerprint)
	if err != nil {
		falseResponse(http.StatusInternalServerError, c)
	} else {
		c.JSON(http.StatusOK, StatusResp{
			Status: true,
		})
	}
}

// DeletePeer delete peer entry.
func (a *APIService) DeletePeer(c *gin.Context) {
	fingerprint := c.Param("fingerprint")
	err := a.tab.Delete(fingerprint)
	if err != nil {
		falseResponse(http.StatusInternalServerError, c)
		return
	}
	c.JSON(http.StatusOK, StatusResp{
		Status: true,
	})
}

// GetProvider returns all peers under the same provider.
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
	var status int
	if id == "" {
		status = http.StatusNotFound
	} else {
		status = http.StatusOK
	}
	c.JSON(status, IDResp{
		PeerID: id,
	})
}

// SetServerID sets server libp2p host id.
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

// GetServerID returns server libp2p host id.
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
