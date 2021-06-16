package route

import (
	"fmt"
	"strings"
	"sync"

	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

type Table struct {
	mx           sync.Mutex
	providers    map[string]map[string]peer.AddrInfo
	peers        map[peer.ID]peer.AddrInfo
	fingerprints map[string]peer.ID
}

func NewRouteTable() *Table {
	return &Table{
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

func (t *Table) Find(id peer.ID) (peer.AddrInfo, error) {
	t.mx.Lock()
	defer t.mx.Unlock()
	pi, ok := t.peers[id]
	if !ok {
		return peer.AddrInfo{}, fmt.Errorf("route: not found")
	}
	return pi, nil
}

func (t *Table) Provide(cid string, id peer.ID, addrs, fingerprint string) error {
	t.mx.Lock()
	defer t.mx.Unlock()
	pmap, ok := t.providers[cid]
	if !ok {
		pmap = make(map[string]peer.AddrInfo)
		t.providers[cid] = pmap
	}

	pi, err := parseAddrInfoHack(id, addrs)
	if err != nil {
		return err
	}
	// If we use peer.ID as map key, json.Marshal can not encode properly,
	// so we save it as string.
	idStr := id.String()
	pmap[idStr] = pi

	t.fingerprints[fingerprint] = id

	if t.peers == nil {
		t.peers = make(map[peer.ID]peer.AddrInfo)
	}
	t.peers[id] = pi

	return nil
}

func (t *Table) FindProvider(provider string) (map[string]peer.AddrInfo, error) {
	pmap, ok := t.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider not found")
	}
	return pmap, nil
}

func (t *Table) FindPeerID(fingerprint string) peer.ID {
	id, ok := t.fingerprints[fingerprint]
	if !ok {
		return ""
	}
	return id
}
