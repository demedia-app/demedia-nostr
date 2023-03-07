package hashutil

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nbd-wtf/go-nostr"
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

func GetSha256(p []byte) []byte {
	h := sha256.New()
	h.Write(p)
	return h.Sum(nil)
}

func StringifyEvent(evt *nostr.Event) string {
	sx := make([]string, len(evt.Tags))
	for i, tag := range evt.Tags {
		sx[i] = fmt.Sprintf("[\"%s\"]", strings.Join(tag, "\",\""))
	}
	return fmt.Sprintf("[0,\"%s\",%d,%d,[%s],\"%s\"]",
		evt.PubKey,
		evt.CreatedAt.Unix(),
		evt.Kind,
		strings.Join(sx, ","),
		evt.Content,
	)
}
