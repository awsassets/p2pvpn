package route

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"strings"
	"sync"
)

type RouteTable struct {
	mx           sync.Mutex
	providers    map[string]map[string]peer.AddrInfo
	peers        map[peer.ID]peer.AddrInfo
	fingerprints map[string]peer.ID
}

func NewRouteTable() *RouteTable {
	return &RouteTable{
		providers:    make(map[string]map[string]peer.AddrInfo),
		fingerprints: make(map[string]peer.ID),
	}
}

// parseAddrInfoHack parses addrs string to peer.AddrInfo
// addrs format: addr,addr
func parseAddrInfoHack(id peer.ID, addrs string) (peer.AddrInfo, error) {
	var remoteAddresses []ma.Multiaddr
	array := strings.Split(addrs, ",")
	for _, addr := range array {
		if addr != "" {
			maAddr, err := ma.NewMultiaddr(addr)
			if err == nil {
				remoteAddresses = append(remoteAddresses, maAddr)
			} else {
				return peer.AddrInfo{}, err
			}
		}
	}
	return peer.AddrInfo{
		ID:    id,
		Addrs: remoteAddresses,
	}, nil

}

func (r *RouteTable) Find(id peer.ID) (peer.AddrInfo, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	pi, ok := r.peers[id]
	if !ok {
		return peer.AddrInfo{}, fmt.Errorf("route: not found")
	}
	return pi, nil
}

func (r *RouteTable) Provide(cid string, id peer.ID, addrs, fingerprint string) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	pmap, ok := r.providers[cid]
	if !ok {
		pmap = make(map[string]peer.AddrInfo)
		r.providers[cid] = pmap
	}

	pi, err := parseAddrInfoHack(id, addrs)
	if err != nil {
		return err
	}
	// If we use peer.ID as map key, json.Marshal can not encode properly,
	// so we save it as string.
	idStr := id.String()
	pmap[idStr] = pi

	r.fingerprints[fingerprint] = id

	if r.peers == nil {
		r.peers = make(map[peer.ID]peer.AddrInfo)
	}
	r.peers[id] = pi

	return nil
}

func (r *RouteTable) FindProvider(provider string) (map[string]peer.AddrInfo, error) {
	pmap, ok := r.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider not found")
	}
	return pmap, nil
}

func (r *RouteTable) FindPeerID(fingerprint string) peer.ID {
	id, ok := r.fingerprints[fingerprint]
	if !ok {
		return ""
	}
	return id
}
