package exporter

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/willglynn/miso_exporter/miso"
)

type renewableProduction struct {
	client *miso.Client

	renewableProduction *prometheus.Desc
	source              miso.Renewable
}

func (c renewableProduction) Describe(descs chan<- *prometheus.Desc) {
	descs <- c.renewableProduction
}

func (c renewableProduction) Collect(metrics chan<- prometheus.Metric) {
	production, err := c.client.RenewableProduction(context.Background(), c.source)
	if err != nil {
		metrics <- prometheus.NewInvalidMetric(c.renewableProduction, err)
		return
	}

	for _, entry := range production {
		m := prometheus.MustNewConstMetric(c.renewableProduction, prometheus.GaugeValue, float64(entry.Megawatts)*1_000_000)
		for t := entry.StartAt; t.Before(entry.EndAt); t = t.Add(time.Minute) {
			metrics <- prometheus.NewMetricWithTimestamp(t, m)
		}
	}
}

func NewRenewableProduction(client *miso.Client, source miso.Renewable) prometheus.Collector {
	return &renewableProduction{
		client: client,
		source: source,

		renewableProduction: prometheus.NewDesc("miso_renewable_production_w",
			"The amount of power produced from renewable sources",
			nil, map[string]string{
				"kind":   "actual",
				"source": source.String(),
			},
		),
	}
}
