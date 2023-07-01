package crypto

import (
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/store/log"
	"testing"
)

func TestPrivateKeySigner_AccountKey(t *testing.T) {
	signer := NewSigner(log.Default())
	content, err := signer.AccountKey("test")
	assert.NoError(t, err)
	assert.Contains(t, content, "public-key-sha3-384: hIedp1AvrWlcDI4uS_qjoFLzjKl5enu4G2FYJpgB3Pj-tUzGlTQBxMBsBmi-tnJR")

}
