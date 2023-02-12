package postgresql

func (b *PostgresBackend) GetPeer(pubkey string) string {
	return b.Map[pubkey].Address
}
