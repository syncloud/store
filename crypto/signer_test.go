package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/syncloud/store/log"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func TestPrivateKeySigner_AccountKey(t *testing.T) {
	signer := NewSigner(log.Default(), "old", "")
	content, err := signer.AccountKey("test")
	assert.NoError(t, err)
	assert.Contains(t, content, "public-key-sha3-384: hIedp1AvrWlcDI4uS_qjoFLzjKl5enu4G2FYJpgB3Pj-tUzGlTQBxMBsBmi-tnJR")

}

func TestPrivateKeySigner_NewKeySelected(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	var buf bytes.Buffer
	arm, err := armor.Encode(&buf, "PGP PRIVATE KEY BLOCK", nil)
	assert.NoError(t, err)
	assert.NoError(t, packet.NewRSAPrivateKey(time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC), rsaKey).Serialize(arm))
	assert.NoError(t, arm.Close())

	priv := RSAPrivateKey(rsaKey)
	want := priv.PublicKey().ID()
	signer := NewSigner(log.Default(), "new", buf.String())
	content, err := signer.AccountKey("test")
	assert.NoError(t, err)
	assert.Contains(t, content, "public-key-sha3-384: "+want)
	assert.NotContains(t, content, "hIedp1AvrWlcDI4uS_qjoFLzjKl5enu4G2FYJpgB3Pj-tUzGlTQBxMBsBmi-tnJR")
}
