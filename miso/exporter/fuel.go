package exporter

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/willglynn/miso_exporter/miso"
)

type fuel struct {
	client *miso.Client

	fuel *prometheus.Desc
}

func (c fuel) Describe(descs chan<- *prometheus.Desc) {
	descs <- c.fuel
}

func (c fuel) Collect(metrics chan<- prometheus.Metric) {
	fuel, err := c.client.Fuel(context.Background())
	if err != nil {
		metrics <- prometheus.NewInvalidMetric(c.fuel, err)
		return
	}

	for _, entry := range fuel {
		m := prometheus.MustNewConstMetric(c.fuel, prometheus.GaugeValue, float64(entry.Megawatts)*1_000_000, entry.Name)
		for t := entry.StartAt; t.Before(entry.EndAt); t = t.Add(time.Minute) {
			metrics <- prometheus.NewMetricWithTimestamp(t, m)
		}
	}
}

func NewFuel(client *miso.Client) prometheus.Collector {
	return &fuel{
		client: client,

		fuel: prometheus.NewDesc("miso_fuel_w",
			"The amount of power produced using a particular fuel",
			[]string{"name"},
			map[string]string{
				"kind": "actual",
			},
		),
	}
}
