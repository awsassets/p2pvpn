package route

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/lp2p/p2pvpn/common/utils"
	"github.com/lp2p/p2pvpn/constant"
	"github.com/lp2p/p2pvpn/log"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// httpClient uses to do send requests.
var httpClient = &http.Client{}

// Route is a implement of PeerRouting and ContentRouting.
type Route struct {
	h      host.Host
	server string
}

// StatusResp receives server response status.
type StatusResp struct {
	Status bool `json:"status"`
}

// ProvidersResp receives FindProvidersAsync response.
type ProvidersResp struct {
	StatusResp
	AddrInfos map[string]peer.AddrInfo `json:"addr_infos,omitempty"`
}

// PeerResp receives FindPeer response
type PeerResp struct {
	StatusResp
	AddrInfo peer.AddrInfo `json:"addr_info,omitempty"`
}

// NewRoute creates a new remote routing for client to use.
func NewRoute(h host.Host, server string) *Route {
	return &Route{
		h:      h,
		server: server,
	}
}

func (r *Route) FindPeer(ctx context.Context, p peer.ID) (peer.AddrInfo, error) {
	resp, err := r.get(ctx, r.server+constant.RoutingUrl+p.Pretty())
	if err != nil {
		return peer.AddrInfo{}, err
	}

	res, err := ioutil.ReadAll(resp.Body)
	var respPtr PeerResp
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

func (r *Route) Provide(ctx context.Context, cid cid.Cid, bcast bool) error {
	if !bcast {
		return nil
	}
	var addrs string
	for _, addr := range r.h.Addrs() {
		addrs += addr.String() + ","
	}

	resp, err := r.postForm(ctx, r.server+constant.RoutingUrl+cid.String(),
		url.Values{
			"id":    {r.h.ID().String()},
			"addrs": {addrs},
		})
	if err != nil {
		return err
	}

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var respPtr StatusResp
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

func (r *Route) FindProvidersAsync(ctx context.Context, cid cid.Cid, limit int) <-chan peer.AddrInfo {
	ch := make(chan peer.AddrInfo)
	go func() {
		defer close(ch)
		resp, err := r.get(ctx, r.server+constant.RoutingProviderUrl+cid.String())
		if err != nil {
			log.Errorf("%v", err)
		}

		res, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("%v", err)
		}

		var respPtr ProvidersResp
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

// MakeRouting returns function for libp2p.Routing, it will register node itself
// when create a new node.
func MakeRouting(server, ns string) func(h host.Host) (routing.PeerRouting, error) {
	var router routing.PeerRouting
	return func(h host.Host) (routing.PeerRouting, error) {
		router = NewRoute(h, server)
		contentRouter := router.(routing.ContentRouting)
		// Use Provide to register the node.
		err := contentRouter.Provide(context.Background(), utils.StrToCid(ns), true)
		return router, err
	}
}

// get Sends get request with context.
func (r *Route) get(ctx context.Context, url string) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return httpClient.Do(req)
}

// post Sends post request with context.
func (r *Route) post(ctx context.Context, url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return httpClient.Do(req)
}

// postForm Sends form request with context.
func (r *Route) postForm(ctx context.Context, url string, data url.Values) (resp *http.Response, err error) {
	return r.post(ctx, url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}
