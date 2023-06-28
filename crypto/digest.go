package crypto

import (
	"crypto"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

const (
	hashDigestBufSize = 2 * 1024 * 1024
)

func EncodeDigest(hash crypto.Hash, hashDigest []byte) (string, error) {
	algo := ""
	switch hash {
	case crypto.SHA512:
		algo = "sha512"
	case crypto.SHA3_384:
		algo = "sha3-384"
	default:
		return "", fmt.Errorf("unsupported hash")
	}
	if len(hashDigest) != hash.Size() {
		return "", fmt.Errorf("hash digest by %s should be %d bytes", algo, hash.Size())
	}
	return base64.RawURLEncoding.EncodeToString(hashDigest), nil
}

func SnapFileSHA3_384(snapPath string) (digest string, size uint64, err error) {
	sha3_384Dgst, size, err := FileDigest(snapPath, crypto.SHA3_384)
	if err != nil {
		return "", 0, fmt.Errorf("cannot compute snap %q digest: %v", snapPath, err)
	}

	sha3_384, err := EncodeDigest(crypto.SHA3_384, sha3_384Dgst)
	if err != nil {
		return "", 0, fmt.Errorf("cannot encode snap %q digest: %v", snapPath, err)
	}
	return sha3_384, size, nil
}

func FileDigest(filename string, hash crypto.Hash) ([]byte, uint64, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	h := hash.New()
	size, err := io.CopyBuffer(h, f, make([]byte, hashDigestBufSize))
	if err != nil {
		return nil, 0, err
	}
	return h.Sum(nil), uint64(size), nil
}
