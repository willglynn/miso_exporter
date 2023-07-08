package exporter

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/willglynn/miso_exporter/miso"
)

type load struct {
	client *miso.Client

	load *prometheus.Desc
}

func (l load) Describe(descs chan<- *prometheus.Desc) {
	descs <- l.load
}

func (l load) Collect(metrics chan<- prometheus.Metric) {
	forecast, err := l.client.LoadAndForecast(context.Background())
	if err != nil {
		metrics <- prometheus.NewInvalidMetric(l.load, err)
		return
	}

	for _, hour := range forecast.HourlyForecast {
		m := prometheus.MustNewConstMetric(l.load, prometheus.GaugeValue, float64(hour.Megawatts)*1_000_000, "forecast")
		end := hour.At.Add(time.Hour)
		for t := hour.At; t.Before(end); t = t.Add(time.Minute) {
			metrics <- prometheus.NewMetricWithTimestamp(t, m)
		}
	}

	for _, min := range forecast.FiveMinuteLoad {
		m := prometheus.MustNewConstMetric(l.load, prometheus.GaugeValue, float64(min.Megawatts)*1_000_000, "actual")
		end := min.At
		for t := min.At.Add(-5 * time.Minute); t.Before(end); t = t.Add(time.Minute) {
			metrics <- prometheus.NewMetricWithTimestamp(t, m)
		}
	}
}

func NewLoad(client *miso.Client) prometheus.Collector {
	return &load{
		client: client,

		load: prometheus.NewDesc("miso_load_total_w",
			"The amount of load, forecast or actual",
			[]string{"kind"}, nil,
		),
	}
}
