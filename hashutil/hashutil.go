package hashutil

import (
	"crypto/ecdsa"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func GetSing(input interface{}, prv *ecdsa.PrivateKey) (string, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	hash := crypto.Keccak256Hash(body)
	sig, err := crypto.Sign(hash.Bytes(), prv)
	if err != nil {
		return "", err
	}
	str := hexutil.Encode(sig)
	return str, nil
}

func GetVerification(encode string, input interface{}, pub *ecdsa.PublicKey) (bool, error) {
	sig, err := hexutil.Decode(encode)
	if err != nil {
		return false, err
	}
	publicKeyBytes := crypto.FromECDSAPub(pub)
	body, err := json.Marshal(input)
	if err != nil {
		return false, err
	}
	hash := crypto.Keccak256Hash(body)
	signatureNoRecoverID := sig[:len(sig)-1] // remove recovery id

	verified := crypto.VerifySignature(publicKeyBytes, hash.Bytes(), signatureNoRecoverID)
	return verified, nil
}
