package core

import (
	"fmt"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/lp2p/p2pvpn/log"
)

// AddAddrToPeerstore parses a peer multiaddress and adds
// it to the given host's peerstore, so it knows how to
// contact it. It returns the peer ID of the remote peer.
func AddAddrToPeerstore(h host.Host, addr string) peer.ID {
	// The following code extracts target's the peer ID from the
	// given multiaddress
	ipfsAddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		log.Errorf("%v", err)
	}
	pid, err := ipfsAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Errorf("%v", err)
	}

	peerID, err := peer.Decode(pid)
	if err != nil {
		log.Errorf("%v", err)
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.Encode(peerID)))
	targetAddr := ipfsAddr.Decapsulate(targetPeerAddr)

	// We have a peer ID and a targetAddr so we add
	// it to the peerstore so LibP2P knows how to contact it
	h.Peerstore().AddAddr(peerID, targetAddr, peerstore.PermanentAddrTTL)
	return peerID
}
