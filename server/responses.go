package server

import "github.com/libp2p/go-libp2p-core/peer"

// StatusResp receives server response status.
type StatusResp struct {
	Status bool `json:"status"`
}

// ProvidersResp receives FindProvidersAsync response.
type ProvidersResp struct {
	Status    bool                     `json:"status"`
	AddrInfos map[string]peer.AddrInfo `json:"addr_infos,omitempty"`
}

// PeerResp receives FindPeer response.
type PeerResp struct {
	Status   bool          `json:"status"`
	AddrInfo peer.AddrInfo `json:"addr_info,omitempty"`
}

// IDResp receives FindPeerID and GetServerID response.
type IDResp struct {
	PeerID peer.ID `json:"peer_id,omitempty"`
}
