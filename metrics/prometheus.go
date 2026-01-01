package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	DbQueryLatencySeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_db_query_latency_seconds",
			Help:    "Latency of database queries in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"query_name"},
	)

	DbQueryTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_db_requests_total",
			Help: "Total number of database requests",
		},
		[]string{"query_name"},
	)

	DbErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_db_errors_total",
			Help: "Total number of database request errors",
		},
		[]string{"query_name"},
	)
)

func Init(reg prometheus.Registerer) {
	reg.MustRegister(DbQueryLatencySeconds)
	reg.MustRegister(DbQueryTotal)
	reg.MustRegister(DbErrorsTotal)
}
