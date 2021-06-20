package route

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lp2p/p2pvpn/server"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/libp2p/go-libp2p/p2p/host/relay"
	"github.com/lp2p/p2pvpn/common/utils"
	"github.com/lp2p/p2pvpn/constant"
	"github.com/lp2p/p2pvpn/log"
)

// httpClient uses to do send requests.
var httpClient = &http.Client{}

// Route is a implement of PeerRouting and ContentRouting.
type Route struct {
	h           host.Host
	fingerprint string
	serverUrl   string
	secret      string
}

var (
	_router *Route
)

func Router() *Route {
	return _router
}

// NewRoute creates a new remote routing for client to use.
func NewRoute(h host.Host, serverUrl, fingerprint, secret string) *Route {
	return &Route{
		h:           h,
		serverUrl:   serverUrl,
		fingerprint: fingerprint,
		secret:      secret,
	}
}

// FindPeer implements routing.PeerRouting.
func (r *Route) FindPeer(ctx context.Context, p peer.ID) (peer.AddrInfo, error) {
	resp, err := r.getWithContext(ctx, r.serverUrl+constant.RoutingUrl+p.Pretty())
	if err != nil {
		return peer.AddrInfo{}, err
	}

	res, err := io.ReadAll(resp.Body)
	var respPtr server.PeerResp
	err = json.Unmarshal(res, &respPtr)
	if err != nil {
		return peer.AddrInfo{}, err
	}

	if respPtr.Status {
		return respPtr.AddrInfo, nil
	} else {
		return peer.AddrInfo{}, nil
	}
}

// Provide implements routing.ContentRouting.
func (r *Route) Provide(ctx context.Context, cid cid.Cid, bcast bool) error {
	if !bcast {
		return nil
	}
	var addrs string
	for _, addr := range r.h.Addrs() {
		addrs += addr.String() + ","
	}

	resp, err := r.postForm(ctx, r.serverUrl+constant.RoutingUrl+cid.String(),
		url.Values{
			"id":          {r.h.ID().String()},
			"addrs":       {addrs},
			"fingerprint": {r.fingerprint},
		})
	if err != nil {
		return err
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var respPtr server.StatusResp
	err = json.Unmarshal(res, &respPtr)
	if err != nil {
		return err
	}

	if respPtr.Status {
		return nil
	} else {
		return fmt.Errorf("something wrong")
	}
}

// FindProvidersAsync implements routing.ContentRouting.
func (r *Route) FindProvidersAsync(ctx context.Context, cid cid.Cid, limit int) <-chan peer.AddrInfo {
	ch := make(chan peer.AddrInfo)
	go func() {
		defer close(ch)
		resp, err := r.getWithContext(ctx, r.serverUrl+constant.RoutingProviderUrl+cid.String())
		if err != nil {
			log.Errorf("%v", err)
		}

		res, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("%v", err)
		}

		var respPtr server.ProvidersResp
		err = json.Unmarshal(res, &respPtr)
		if err != nil {
			log.Errorf("%v", err)
		}

		for _, pi := range respPtr.AddrInfos {
			select {
			case ch <- pi:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}

// FindPeerID finds peer id by fingerprint.
func (r *Route) FindPeerID(fingerprint string) (peer.ID, error) {
	resp, err := r.get(r.serverUrl + constant.FingerprintsUrl + fingerprint)
	if err != nil {
		return "", err
	}

	res, err := io.ReadAll(resp.Body)
	var respPtr server.IDResp
	err = json.Unmarshal(res, &respPtr)
	if err != nil {
		return "", err
	}
	return respPtr.PeerID, nil
}

func (r *Route) Logout(fingerprint string) error {
	resp, err := r.delete(r.serverUrl + constant.FingerprintsUrl + fingerprint)
	if err != nil {
		return err
	}

	res, err := io.ReadAll(resp.Body)
	var respPtr server.StatusResp
	err = json.Unmarshal(res, &respPtr)
	if err != nil {
		return err
	}
	if !respPtr.Status {
		return fmt.Errorf("fail to delete")
	}
	return nil
}

func (r *Route) SetServerID() error {
	resp, err := r.post(r.serverUrl+constant.ServerIDUrl+r.h.ID().String(), "", nil)
	if err != nil {
		return err
	}
	res, err := io.ReadAll(resp.Body)
	var respPtr server.StatusResp
	err = json.Unmarshal(res, &respPtr)
	if err != nil {
		return err
	}
	if !respPtr.Status {
		return fmt.Errorf("failed to register server id")
	}
	return nil
}

func (r *Route) GetServerID() (peer.ID, error) {
	resp, err := r.get(r.serverUrl + constant.ServerIDUrl)
	if err != nil {
		return "", err
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var respPtr server.IDResp
	err = json.Unmarshal(res, &respPtr)
	if err != nil {
		return "", err
	}
	return respPtr.PeerID, nil
}

// MakeRouting returns function for libp2p.Routing, it will register node itself
// when create a new node.
func MakeRouting(serverUrl, ns, fingerprint, secret string) func(h host.Host) (routing.PeerRouting, error) {
	var router routing.PeerRouting
	return func(h host.Host) (routing.PeerRouting, error) {
		_router = NewRoute(h, serverUrl, fingerprint, secret)
		router = _router
		var err error
		// Only register ourself when namespace is not relay.RelayRendezvous
		if ns != relay.RelayRendezvous {
			contentRouter := router.(routing.ContentRouting)
			// Use Provide to register the node.
			err = contentRouter.Provide(context.Background(), utils.StrToCid(ns), true)
		}
		return router, err
	}
}

// get sends get request with auth header.
func (r *Route) get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	r.setAuthHeader(req)
	return httpClient.Do(req)
}

func (r *Route) delete(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return
	}
	r.setAuthHeader(req)
	return httpClient.Do(req)
}

// getWithContext sends get request with auth header and context.
func (r *Route) getWithContext(ctx context.Context, url string) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	r.setAuthHeader(req)
	return httpClient.Do(req)
}

// post sends post request with auth header.
func (r *Route) post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return
	}
	r.setAuthHeader(req)
	req.Header.Set("Content-Type", contentType)
	return httpClient.Do(req)
}

// postWithContext sends post request with context.
func (r *Route) postWithContext(ctx context.Context, url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return
	}
	r.setAuthHeader(req)
	req.Header.Set("Content-Type", contentType)
	return httpClient.Do(req)
}

// postForm sends form request with auth header and context.
func (r *Route) postForm(ctx context.Context, url string, data url.Values) (resp *http.Response, err error) {
	return r.postWithContext(ctx, url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (r *Route) setAuthHeader(req *http.Request) {
	secret := utils.Md5(r.secret)
	req.Header.Set("auth", secret)
}
