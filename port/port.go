package port

import (
	"math/rand"
	"time"
)

func GetTargetAddressPort() int {
	rand.Seed(time.Now().UnixNano())
	port := rand.Intn(1000) + 10000
	return port
}
