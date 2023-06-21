package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	gorpc "github.com/libp2p/go-libp2p-gorpc"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sithumonline/demedia-nostr/blob"
	p2pHost "github.com/sithumonline/demedia-nostr/host"
	"github.com/sithumonline/demedia-nostr/hub/handler"
	"github.com/sithumonline/demedia-nostr/hub/ping"
	"github.com/sithumonline/demedia-nostr/keys"
	"github.com/sithumonline/demedia-nostr/relayer"
	"github.com/sithumonline/demedia-nostr/relayer/storage/postgresql"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type Relay struct {
	PostgresDatabase string `envconfig:"POSTGRESQL_DATABASE"`

	storage *postgresql.PostgresBackend

	Hex string `envconfig:"HEX" default:"fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19"`

	BlobID string `envconfig:"BLOB_ID" default:"hub"`

	BucketURI string `envconfig:"BUCKET_URI"`

	WebPort string `envconfig:"WEB_PORT" default:"3030"`

	P2PPort string `envconfig:"P2P_PORT" default:"10880"`

	RelayPort string `envconfig:"RELAY_PORT" default:"7448"`

	LocalNet string `envconfig:"LOCAL_NET" default:"1"`

	Environment string `envconfig:"ENVIRONMENT" default:"development"`

	Version string `envconfig:"VERSION" default:"0.0.1"`
}

func (r *Relay) Name() string {
	return "Hub"
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
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			for k, e := range r.storage.Map {
				ts := time.Now().Sub(e.LastUpdate)
				tg := 5 * time.Second
				if ts > tg {
					delete(r.storage.Map, k)
				}
			}
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
	tracer.Start(tracer.WithServiceName(r.Name()), tracer.WithEnv(r.Environment), tracer.WithServiceVersion(r.Version))
	defer tracer.Stop()
	r.storage = &postgresql.PostgresBackend{
		DatabaseURL: r.PostgresDatabase,
		Map:         map[string]postgresql.PeerInfo{},
		ServiceName: r.Name(),
	}
	ecdsaPvtKey, privKey, _, _, err := keys.GetKeys(r.Hex)
	if err != nil {
		log.Fatalf("failed to get priv key for libp2p: %v", err)
	}
	add := p2pHost.GetAdd(r.P2PPort, r.LocalNet)
	h, err := p2pHost.GetHost(*privKey, add)
	if err != nil {
		log.Fatalf("failed to get host: %v", err)
	}
	hostAddr := p2pHost.GetMultiAddr(h)
	log.Printf("Hub: listening on %s\n", hostAddr)
	rpcHost := gorpc.NewServer(h, "/p2p/1.0.0")
	pingService := ping.NewPingService(&r)
	if err := rpcHost.Register(pingService); err != nil {
		log.Fatalf("failed to register rpc server: %v", err)
	}
	rs := relayer.Settings{Port: r.RelayPort}
	cfg := blob.AuditTrail{
		ID:        r.BlobID,
		BucketURI: r.BucketURI,
	}
	b, err := blob.NewBlobStorage(&cfg)
	if err != nil {
		log.Fatalf("failed to up blob: %v", err)
	}
	go handler.Start(fmt.Sprintf(":%s", r.WebPort), r.storage.Map)
	if err := relayer.StartConf(rs, &r, h, b, ecdsaPvtKey); err != nil {
		log.Fatalf("server terminated: %v", err)
	}
}
