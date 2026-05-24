package test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestPublishedApps(t *testing.T) {
	resp, err := resty.New().R().Get("http://api.store/api/ui/v1/apps?channel=stable")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode(), string(resp.Body()))

	body := string(resp.Body())
	assert.Contains(t, body, `"snapId":"testapp1.`, "testapp1 missing from apps list: %s", body)
	assert.Contains(t, body, `"snapId":"testapp2.`, "testapp2 missing from apps list: %s", body)

	// The publisher writes apps/<name>_<version>_<arch>.snap with the version baked
	// into the key. No other test touches these specific objects, so this is robust
	// to test ordering / SetVersion mutations elsewhere.
	for _, key := range []string{
		"apps/testapp1_3_amd64.snap",
		"apps/testapp2_2_amd64.snap",
	} {
		_, err := s3client().HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(Bucket),
			Key:    aws.String(key),
		})
		assert.NoError(t, err, "publisher should have uploaded %s", key)
	}
}
