package host

import (
	"fmt"
	"log"
	"net"

	externalip "github.com/glendc/go-external-ip"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
)

func GetAdd(port string, isLocal string) string {
	consensus := externalip.DefaultConsensus(nil, nil)
	parsedIP, err := consensus.ExternalIP()
	if err != nil {
		log.Panic(fmt.Errorf("get external ip error: %v", err))
	}
	if isLocal == "1" {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			log.Panic(fmt.Errorf("get local ip error: %v", err))
		}
		for _, address := range addrs {
			// check the address type and if it is not a loopback the display it
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				parsedIP = ipnet.IP
				break
			}
		}
	}
	if err != nil {
		log.Panic(err)
	}
	if parsedIP.To4() != nil {
		return fmt.Sprintf("/ip4/%s/tcp/%s", parsedIP.String(), port)
	} else {
		return fmt.Sprintf("/ip6/%s/tcp/%s", parsedIP.String(), port)
	}
}

func GetHost(prvKey crypto.PrivKey, add string) (host.Host, error) {
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(add),
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
