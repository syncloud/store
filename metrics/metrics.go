package metrics

import "github.com/prometheus/client_golang/prometheus"

var PopularityRecord = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "store_popularity_record_total",
		Help: "Number of device check-ins recorded, by snap.",
	},
	[]string{"snap"},
)

func init() {
	prometheus.MustRegister(PopularityRecord)
}
