package api

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestScheduledRefreshCounter(t *testing.T) {
	m := NewSnapdMetrics()
	m.RecordScheduledRefresh("2.59.5")
	m.RecordScheduledRefresh("2.59.5")
	m.RecordScheduledRefresh("2.61")

	if got := testutil.ToFloat64(m.scheduledRefreshes.WithLabelValues("2.59.5")); got != 2 {
		t.Errorf("scheduled refresh 2.59.5 = %v, want 2", got)
	}
	if got := testutil.ToFloat64(m.scheduledRefreshes.WithLabelValues("2.61")); got != 1 {
		t.Errorf("scheduled refresh 2.61 = %v, want 1", got)
	}
}
