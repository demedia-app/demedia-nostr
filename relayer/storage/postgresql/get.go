package postgresql

import (
	"math/rand"
	"time"
)

func (b *PostgresBackend) GetPeer(pubkey string) string {
	address := b.Map[pubkey].Address
	if address == "" {
		if len(b.Map) == 0 {
			return ""
		}
		rand.Seed(time.Now().UnixNano())
		n := rand.Intn(len(b.Map))
		i := 0
		for _, v := range b.Map {
			if i == n {
				return v.Address
			}
			i++
		}
	}

	return address
}
