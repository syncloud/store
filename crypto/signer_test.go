package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syncloud/store/log"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

var fixedTs = time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC)

func armorKey(t *testing.T, rsaKey *rsa.PrivateKey) string {
	var buf bytes.Buffer
	arm, err := armor.Encode(&buf, "PGP PRIVATE KEY BLOCK", nil)
	require.NoError(t, err)
	require.NoError(t, packet.NewRSAPrivateKey(fixedTs, rsaKey).Serialize(arm))
	require.NoError(t, arm.Close())
	return buf.String()
}

// honestVerify checks the assertion's signature the way real snapd does
// (raw openpgp), independent of the fork's disabled SignatureCheck. The signed
// content is everything before the final "\n\n"; the signature follows it.
func honestVerify(t *testing.T, assertionText string, pub *rsa.PublicKey) {
	idx := strings.LastIndex(assertionText, "\n\n")
	require.GreaterOrEqual(t, idx, 0)
	content := assertionText[:idx]
	sigB64 := strings.ReplaceAll(strings.TrimRight(assertionText[idx+2:], "\n"), "\n", "")
	raw, err := base64.StdEncoding.DecodeString(sigB64)
	require.NoError(t, err)
	require.Equal(t, byte(0x01), raw[0])
	pkt, err := packet.Read(bytes.NewReader(raw[1:]))
	require.NoError(t, err)
	sig := pkt.(*packet.Signature)
	pubPkt := packet.NewRSAPublicKey(fixedTs, pub)
	h := sig.Hash.New()
	h.Write([]byte(content))
	assert.NoError(t, pubPkt.VerifySignature(h, sig))
}

func assertSignerVerifies(t *testing.T, signer *PrivateKeySigner, pub *rsa.PublicKey) {
	ak, err := signer.AccountKey("test")
	require.NoError(t, err)
	honestVerify(t, ak, pub)

	sd, err := signer.SnapDeclaration("16", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	require.NoError(t, err)
	honestVerify(t, sd, pub)

	rev := `{"snap-revision":"7","snap-id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","snap-size":"1024"}`
	sr, err := signer.SnapRevision("NX7IJUakITdsfLFyqWHkdS-YTp77S-v2-vckEZ1cVBwElYYld8_34m6Ni9JZ1aQL", rev)
	require.NoError(t, err)
	honestVerify(t, sr, pub)
}

func TestSigner_AssertionsVerifyHonestly_OldKey(t *testing.T) {
	_, oldRsa := ReadPrivateKey(syncloudPrivKey)
	assertSignerVerifies(t, NewSigner(log.Default(), "old", ""), &oldRsa.PublicKey)
}

func TestSigner_AssertionsVerifyHonestly_NewKey(t *testing.T) {
	newRsa, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	assertSignerVerifies(t, NewSigner(log.Default(), "new", armorKey(t, newRsa)), &newRsa.PublicKey)
}

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
