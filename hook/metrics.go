package hook

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/dwellersclub/contigus/utils"
)

//Metrics Metrics for hooks
type Metrics interface {
	IncHandled(hookType string, ID string, errorCode string, elapseTime time.Duration)
}

// NewHookMetrics Creates a new metrics for hooks
func NewHookMetrics() Metrics {

	requestTotal := make(map[string]utils.Counter)

	hookTypes := append([]string{""}, Types.Values()...)
	errorCodes := append([]string{""}, Errors.Values()...)

	for _, hookType := range hookTypes {
		for _, errorCode := range errorCodes {
			key := fmt.Sprintf("%s_%s", hookType, errorCode)

			requestCounter := utils.Counter(int32(0))

			counter := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
				Namespace:   "contigus",
				Subsystem:   "hook",
				Name:        "request_total",
				Help:        "Total number of hook rejected for a given type",
				ConstLabels: map[string]string{"type": hookType, "error": errorCode},
			}, requestCounter.Reset)

			prometheus.MustRegister(counter)

			requestTotal[key] = requestCounter
		}
	}

	duration := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "request_duration_seconds",
		Help:       "Time taken to handle a hook",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"type", "id"})

	return &hookPrometheusMetrics{
		requestTotal: requestTotal,
		duration:     duration,
	}
}

// hookPrometheusMetrics Metrics implementation for parser with Promotheus
type hookPrometheusMetrics struct {
	requestTotal map[string]utils.Counter
	duration     *prometheus.SummaryVec
}

func (hp *hookPrometheusMetrics) IncHandled(hookType string, ID string, errorCode string, elapseTime time.Duration) {
	key := fmt.Sprintf("%s_%s", hookType, errorCode)
	counter, ok := hp.requestTotal[key]
	if !ok {
		// log error
		logrus.Errorf("No metrics with keys %s", key)
		return
	}
	counter.Inc()

	if len(errorCode) == 0 {
		hp.duration.WithLabelValues(hookType, ID).Observe(elapseTime.Seconds())
	}
}
