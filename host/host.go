package host

import (
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
	"log"
)

func GetHost(port int, prvKey crypto.PrivKey) (host.Host, error) {
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port+1)),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func GetMultiAddr(h host.Host) multiaddr.Multiaddr {
	addr := h.Addrs()[0]
	ipfsAddr, err := multiaddr.NewMultiaddr("/ipfs/" + h.ID().String())
	if err != nil {
		log.Panic(err)
	}
	multiAddr := addr.Encapsulate(ipfsAddr)
	return multiAddr
}
