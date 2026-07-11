package api

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
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

func TestActiveDevicesGauge(t *testing.T) {
	now := time.Unix(0, 0)
	m := NewSnapdMetrics()
	m.now = func() time.Time { return now }

	m.RecordScheduledRefresh("2.59.5")
	m.RecordScheduledRefresh("2.59.5")
	m.RecordScheduledRefresh("2.61")

	if got := gaugeValue(t, m, "2.59.5"); got != 2 {
		t.Errorf("active devices 2.59.5 = %v, want 2", got)
	}
	if got := gaugeValue(t, m, "2.61"); got != 1 {
		t.Errorf("active devices 2.61 = %v, want 1", got)
	}

	now = now.Add(activeDevicesWindow + time.Minute)
	m.RecordScheduledRefresh("2.61")

	if got := gaugeValue(t, m, "2.61"); got != 1 {
		t.Errorf("active devices 2.61 after window = %v, want 1", got)
	}
	if got := gaugeValue(t, m, "2.59.5"); got != 0 {
		t.Errorf("active devices 2.59.5 after window = %v, want 0 (dropped out)", got)
	}
}

func gaugeValue(t *testing.T, m *SnapdMetrics, version string) float64 {
	t.Helper()
	ch := make(chan prometheus.Metric, 16)
	m.Collect(ch)
	close(ch)
	for metric := range ch {
		var pb dto.Metric
		if err := metric.Write(&pb); err != nil {
			t.Fatal(err)
		}
		if pb.Gauge == nil {
			continue
		}
		for _, label := range pb.GetLabel() {
			if label.GetName() == "version" && label.GetValue() == version {
				return pb.GetGauge().GetValue()
			}
		}
	}
	return 0
}
