package route

import (
	"fmt"
	ma "github.com/multiformats/go-multiaddr"
	"strings"
	"sync"
)

// AddrInfo is fork from peer.AddrInfo, create this for server
// response ID without encode by base58
type AddrInfo struct {
	ID    string
	Addrs []ma.Multiaddr
}

type RouteTable struct {
	mx        sync.Mutex
	providers map[string]map[string]AddrInfo
	peers     map[string]AddrInfo
}

func NewRouteTable() *RouteTable {
	return &RouteTable{providers: make(map[string]map[string]AddrInfo)}
}

// parseAddrInfoHack parses addrs string to AddrInfo
// addrs format: addr,addr
func parseAddrInfoHack(id string, addrs string) (AddrInfo, error) {
	var remoteAddresses []ma.Multiaddr
	array := strings.Split(addrs, ",")
	for _, addr := range array {
		if addr != "" {
			maAddr, err := ma.NewMultiaddr(addr)
			if err == nil {
				remoteAddresses = append(remoteAddresses, maAddr)
			} else {
				return AddrInfo{}, err
			}
		}
	}
	return AddrInfo{
		ID:    id,
		Addrs: remoteAddresses,
	}, nil

}

func (r *RouteTable) Find(id string) (AddrInfo, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	pi, ok := r.peers[id]
	if !ok {
		return AddrInfo{}, fmt.Errorf("route: not found")
	}
	return pi, nil
}

func (r *RouteTable) Provide(cid, id, addrs string) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	pmap, ok := r.providers[cid]
	if !ok {
		pmap = make(map[string]AddrInfo)
		r.providers[cid] = pmap
	}

	pi, err := parseAddrInfoHack(id, addrs)
	if err != nil {
		return err
	}
	pmap[id] = pi
	if r.peers == nil {
		r.peers = make(map[string]AddrInfo)
	}
	r.peers[id] = pi

	return nil
}

func (r *RouteTable) FindProvider(provider string) (map[string]AddrInfo, error) {
	pmap, ok := r.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider not found")
	}
	return pmap, nil
}
