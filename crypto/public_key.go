package crypto

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"fmt"
	"golang.org/x/crypto/openpgp/packet"
	"golang.org/x/crypto/sha3"
	"time"
)

const (
	maxEncodeLineLength = 76
	v1                  = 0x1
)

var (
	v1Header         = []byte{v1}
	v1FixedTimestamp = time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC)
)

type OpenpgpPubKey struct {
	pubKey   *packet.PublicKey
	sha3_384 string
}

func (k *OpenpgpPubKey) ID() string {
	return k.sha3_384
}

func newOpenPGPPubKey(intPubKey *packet.PublicKey) *OpenpgpPubKey {
	h := sha3.New384()
	h.Write(v1Header)
	err := intPubKey.Serialize(h)
	if err != nil {
		panic("internal error: cannot compute public key sha3-384")
	}
	sha3_384, err := EncodeDigest(crypto.SHA3_384, h.Sum(nil))
	if err != nil {
		panic("internal error: cannot compute public key sha3-384")
	}
	return &OpenpgpPubKey{pubKey: intPubKey, sha3_384: sha3_384}
}

func RSAPrivateKey(privateKey *rsa.PrivateKey) OpenpgpPrivateKey {
	pgpPrivateKey := packet.NewRSAPrivateKey(v1FixedTimestamp, privateKey)
	return OpenpgpPrivateKey{pgpPrivateKey}
}

func (k *OpenpgpPubKey) EncodeKey() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := k.pubKey.Serialize(buf)
	if err != nil {
		return nil, fmt.Errorf("cannot encode public key: %v", err)
	}
	return EncodeV1(buf.Bytes()), nil
}
