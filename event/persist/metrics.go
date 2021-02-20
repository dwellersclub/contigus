package persist

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// NewPersisterMetrics Creates a new metrics for Persister
func NewPersisterMetrics() PersisterMetrics {

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "marky",
			Subsystem: "persister",
			Name:      "persisted_total",
			Help:      "Total number of items persisted",
		},
		[]string{"type", "success"},
	)

	timer := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "marky",
			Subsystem: "persister",
			Name:      "persisted_second",
			Help:      "Total time taken to persist items",
		},
		[]string{"type", "success"},
	)

	prometheus.MustRegister(timer)
	prometheus.MustRegister(counter)

	return &PersisterPrometheusMetrics{execCounter: counter, execTimeSummary: timer}
}

// PersisterMetrics Metrics interface for each Persister
type PersisterMetrics interface {
	IncPersist(itemCount int, success bool, persistType string)
	IncPersistTime(duration time.Duration, success bool, persistType string)
}

// PersisterPrometheusMetrics Metrics implementation for parser with Promotheus
type PersisterPrometheusMetrics struct {
	execCounter     *prometheus.CounterVec
	execTimeSummary *prometheus.SummaryVec
	site            string
}

// IncPersist counter to record how many time a Persister was called
func (pg *PersisterPrometheusMetrics) IncPersist(itemCount int, success bool, persistType string) {
	pg.execCounter.WithLabelValues(persistType, strconv.FormatBool(success)).Add(float64(itemCount))
}

// IncPersistTime Record time taken by request
func (pg *PersisterPrometheusMetrics) IncPersistTime(duration time.Duration, success bool, persistType string) {
	pg.execTimeSummary.WithLabelValues(persistType, strconv.FormatBool(success)).Observe(duration.Seconds())
}
