package keys

import (
	"crypto/ecdsa"

	"github.com/btcsuite/btcd/btcec/v2"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p/core/crypto"
)

func GetKeys(hex string) (*ecdsa.PrivateKey, *crypto.PrivKey, *btcec.PrivateKey, *btcec.PublicKey, error) {
	var ecdsaPvtKey *ecdsa.PrivateKey
	var err error
	if hex != "" {
		ecdsaPvtKey, err = ethcrypto.HexToECDSA(hex)
		if err != nil {
			return nil, nil, nil, nil, err
		}
	} else {
		ecdsaPvtKey, err = ethcrypto.GenerateKey()
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}

	libp2pPvtKey, err := crypto.UnmarshalSecp256k1PrivateKey(ecdsaPvtKey.D.Bytes())
	if err != nil {
		return nil, nil, nil, nil, err
	}

	btcecPvtKey, btcecPubKey := btcec.PrivKeyFromBytes(ecdsaPvtKey.D.Bytes())

	return ecdsaPvtKey, &libp2pPvtKey, btcecPvtKey, btcecPubKey, nil
}
