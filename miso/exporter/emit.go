package exporter

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func emit(metrics chan<- prometheus.Metric, m prometheus.Metric, startAt, endAt time.Time) {
	for t := endAt.Add(-time.Second); !t.Before(startAt); t = t.Add(-time.Minute) {
		metrics <- prometheus.NewMetricWithTimestamp(t, m)
	}
}
