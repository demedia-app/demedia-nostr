package postgresql

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type PeerInfo struct {
	Address    string
	LastUpdate time.Time
}

type PostgresBackend struct {
	*sqlx.DB
	DatabaseURL string
	Map         map[string]PeerInfo
	ServiceName string
}
