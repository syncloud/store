package storage

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestPopularity_RecordCountsUniqueDevices(t *testing.T) {
	p := NewPopularity(time.Hour)

	p.Record("a", "d1")
	p.Record("a", "d2")
	p.Record("a", "d1")
	p.Record("b", "d1")

	assert.Equal(t, 2, p.Count("a"))
	assert.Equal(t, 1, p.Count("b"))
	assert.Equal(t, 0, p.Count("c"))
}

func TestPopularity_IgnoresEmpty(t *testing.T) {
	p := NewPopularity(time.Hour)

	p.Record("", "d1")
	p.Record("a", "")

	assert.Equal(t, 0, p.Count("a"))
	assert.Equal(t, 0, p.Count(""))
}

func TestPopularity_CollectExposesGauges(t *testing.T) {
	p := NewPopularity(time.Hour)
	p.Record("a", "d1")
	p.Record("a", "d2")
	p.Record("b", "d1")

	reg := prometheus.NewRegistry()
	reg.MustRegister(p)

	err := testutil.GatherAndCompare(reg, strings.NewReader(`
# HELP store_popularity_devices_active Unique devices active within the TTL window, by snap.
# TYPE store_popularity_devices_active gauge
store_popularity_devices_active{snap="a"} 2
store_popularity_devices_active{snap="b"} 1
# HELP store_popularity_devices_unique Total unique devices active within the TTL window across all snaps.
# TYPE store_popularity_devices_unique gauge
store_popularity_devices_unique 2
# HELP store_popularity_record_total Number of device check-ins recorded, by snap.
# TYPE store_popularity_record_total counter
store_popularity_record_total{snap="a"} 2
store_popularity_record_total{snap="b"} 1
`))
	assert.NoError(t, err)
}
