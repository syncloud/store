package crypto

import (
	"encoding/json"
	"fmt"
	"github.com/syncloud/store/model"
	"go.uber.org/zap"
	"strconv"
	"time"
)

const (
	syncloudPrivKey = `-----BEGIN PGP PRIVATE KEY BLOCK-----
Version: GnuPG v1

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
97WP6Yv2
=gJ0v
-----END PGP PRIVATE KEY BLOCK-----
`
)

type PrivateKeySigner struct {
	privateKey OpenpgpPrivateKey
	logger     *zap.Logger
}

func NewSigner(logger *zap.Logger) *PrivateKeySigner {
	privateKey, _ := ReadPrivateKey(syncloudPrivKey)
	return &PrivateKeySigner{
		privateKey: privateKey,
		logger:     logger,
	}
}

func (s *PrivateKeySigner) SnapRevision(key, revision string) (string, error) {
	var snapRevision model.SnapRevision
	err := json.Unmarshal([]byte(revision), &snapRevision)
	if err != nil {
		return "", err
	}
	headers := "" +
		"snap-revision: " + snapRevision.Revision + "\n" +
		"snap-id: " + snapRevision.Id + "\n" +
		"snap-size: " + snapRevision.Size + "\n" +
		"snap-sha3-384: " + snapRevision.Sha384 + "\n"

	return s.sign("snap-revision", key, headers, "")
}

func (s *PrivateKeySigner) SnapDeclaration(series, snapId string) (string, error) {
	name := model.SnapId(snapId).Name()
	headers := "" +
		"series: " + series + "\n" +
		"snap-id: " + snapId + "\n" +
		"snap-name: " + name + "\n"

	return s.sign("snap-declaration", fmt.Sprintf("%s/%s", series, snapId), headers, "")

}

func (s *PrivateKeySigner) AccountKey(key string) (string, error) {
	publicKeyEnc, err := s.privateKey.PublicKey().EncodeKey()
	if err != nil {
		return "", err
	}

	return s.sign("account-key", key, "", string(publicKeyEnc))
}

func (s *PrivateKeySigner) sign(assertType string, primaryKey string, headers string, body string) (string, error) {
	publicKeyId := s.privateKey.PublicKey().ID()

	s.logger.Info("public key", zap.String("id", publicKeyId))

	content := "type: " + assertType + "\n" +
		"authority-id: syncloud\n" +
		"primary-key: " + primaryKey + "\n" +
		"publisher-id: syncloud\n" +
		"developer-id: syncloud\n" +
		"account-id: syncloud\n" +
		// "display-name: syncloud\n" +
		"revision: 1\n" +
		"sign-key-sha3-384: " + publicKeyId + "\n" +
		"sha3-384: " + publicKeyId + "\n" +
		"public-key-sha3-384: " + publicKeyId + "\n" +
		"timestamp: " + time.Now().Format(time.RFC3339) + "\n" +
		"since: " + time.Now().Format(time.RFC3339) + "\n" +
		headers +
		"validation: certified\n" +
		"body-length: " + strconv.Itoa(len(body)) + "\n\n" +
		body +
		"\n\n"

	fmt.Println("content:")
	fmt.Println(content)
	signature, err := s.privateKey.SignContent([]byte(content))
	if err != nil {
		return "", err
	}
	fmt.Println("signature:")
	fmt.Println(string(signature))
	assertionText := content + string(signature) + "\n"
	return assertionText, nil
}
