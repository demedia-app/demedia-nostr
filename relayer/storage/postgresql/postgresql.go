package postgresql

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type PostgresBackend struct {
	*sqlx.DB
	DatabaseURL string
	Map         map[string]struct {
		Address    string
		LastUpdate time.Time
	}
}
