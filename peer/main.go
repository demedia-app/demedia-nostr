package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/kelseyhightower/envconfig"
	gorpc "github.com/libp2p/go-libp2p-gorpc"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/nbd-wtf/go-nostr"
	p2pHost "github.com/sithumonline/demedia-nostr/host"
	"github.com/sithumonline/demedia-nostr/keys"
	"github.com/sithumonline/demedia-nostr/peer/bridge"
	"github.com/sithumonline/demedia-nostr/peer/handler"
	"github.com/sithumonline/demedia-nostr/port"
	"github.com/sithumonline/demedia-nostr/relayer"
	"github.com/sithumonline/demedia-nostr/relayer/ql"
	"github.com/sithumonline/demedia-nostr/relayer/storage/postgresql"
)

type Relay struct {
	PostgresDatabase string `envconfig:"POSTGRESQL_DATABASE"`

	storage *postgresql.PostgresBackend

	host host.Host

	Hex string `envconfig:"HEX" default:"fad5c8855b840a0b1ed4c323dbad0f11a83a49cad6b3fe8d5819ac83d38b6a19"`

	PeerAddress string

	BtcPubKey string

	Hub string `envconfig:"HUB" default:"/ip4/192.168.1.2/tcp/10880/p2p/16Uiu2HAmP44YB5WWWdYccDYRzByum6fWDma13csdVUcySzwPMqYx"`

	WebPort string `envconfig:"WEB_PORT" default:"4000"`

	P2PPort string `envconfig:"P2P_PORT" default:"10880"`

	LocalNet string `envconfig:"LOCAL_NET" default:"1"`
}

func (r *Relay) Name() string {
	return "Peer"
}

func (r *Relay) Storage() relayer.Storage {
	return r.storage
}

func (r *Relay) OnInitialized(*relayer.Server) {}

func (r *Relay) Init() error {
	err := envconfig.Process("", r)
	if err != nil {
		return fmt.Errorf("couldn't process envconfig: %w", err)
	}

	// every hour, delete all very old events
	go func() {
		db := r.Storage().(*postgresql.PostgresBackend)

		for {
			time.Sleep(60 * time.Minute)
			db.DB.Exec(`DELETE FROM event WHERE created_at < $1`, time.Now().AddDate(0, -3, 0).Unix()) // 3 months
		}
	}()

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		for range ticker.C {
			reply, err := ql.QlCall(r.host, context.Background(), fmt.Sprintf("%s;%s", r.BtcPubKey, r.PeerAddress), r.Hub, "PingService", "Ping", "", "ping-pong")
			if err != nil {
				if strings.Contains(fmt.Sprint(err), "connection refused") {
					log.Println("connection refused, please check the address")
					ticker.Reset(10 * time.Second)
					continue
				} else if strings.Contains(fmt.Sprint(err), "dial backoff") {
					ticker.Reset(15 * time.Second)
					log.Print(err)
					continue
				} else {
					log.Panic(err)
				}
			}
			log.Printf("Response from hub: %s\n", reply.Data)
			ticker.Reset(5 * time.Second)
		}
	}()

	return nil
}

func (r *Relay) AcceptEvent(evt *nostr.Event) bool {
	// block events that are too large
	jsonb, _ := json.Marshal(evt)
	if len(jsonb) > 10000 {
		return false
	}

	return true
}

func main() {
	r := Relay{}
	if err := envconfig.Process("", &r); err != nil {
		log.Fatalf("failed to read from env: %v", err)
	}
	r.storage = &postgresql.PostgresBackend{DatabaseURL: r.PostgresDatabase}
	var p string
	if r.P2PPort == "10880" {
		p = fmt.Sprintf("%d", port.GetTargetAddressPort())
	} else {
		p = r.P2PPort
	}
	_, privKey, btcPvtKey, btcPubKey, err := keys.GetKeys(r.Hex)
	if err != nil {
		log.Fatalf("failed to get priv key for libp2p: %v", err)
	}
	r.BtcPubKey = hex.EncodeToString(schnorr.SerializePubKey(btcPubKey))
	log.Printf("BTC PvtKey: %s", hex.EncodeToString(btcPvtKey.Serialize()))
	add := p2pHost.GetAdd(p, r.LocalNet)
	h, err := p2pHost.GetHost(*privKey, add)
	if err != nil {
		log.Fatalf("failed to get host: %v", err)
	}
	r.host = h
	peerAddr := p2pHost.GetMultiAddr(h)
	r.PeerAddress = peerAddr.String()
	log.Printf("Peer: listening on %s\n", peerAddr)
	rpcHost := gorpc.NewServer(h, "/p2p/1.0.0")
	bridgeService := bridge.NewBridgeService(&r)
	if err := rpcHost.Register(bridgeService); err != nil {
		log.Fatalf("failed to register rpc server: %v", err)
	}
	go handler.Start(fmt.Sprintf(":%s", r.WebPort), &r)
	if err := relayer.Start(&r, nil, nil, nil); err != nil {
		log.Fatalf("server terminated: %v", err)
	}
}
