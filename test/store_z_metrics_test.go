package test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestPopularityMetrics(t *testing.T) {
	client := resty.New()

	type vmResponse struct {
		Data struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	var value int
	for i := 0; i < 30; i++ {
		resp, err := client.R().
			SetQueryParam("query", `store_popularity_record_total{snap="testapp1"}`).
			Get("http://vm:8428/api/v1/query")
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode(), string(resp.Body()))

		var parsed vmResponse
		assert.NoError(t, json.Unmarshal(resp.Body(), &parsed), string(resp.Body()))
		if len(parsed.Data.Result) > 0 {
			var n float64
			if s, ok := parsed.Data.Result[0].Value[1].(string); ok {
				_ = json.Unmarshal([]byte(s), &n)
			}
			value = int(n)
			t.Logf("attempt %d: store_popularity_record_total{snap=testapp1} = %d", i+1, value)
			if value > 0 {
				break
			}
		} else {
			t.Logf("attempt %d: no series yet", i+1)
		}
		time.Sleep(2 * time.Second)
	}
	assert.Greater(t, value, 0, "VM never saw store_popularity_record_total{snap=testapp1}")
}
