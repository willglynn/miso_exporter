package exporter

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/willglynn/miso_exporter/miso"
)

type price struct {
	client *miso.Client

	price *prometheus.Desc
}

func (l price) Describe(descs chan<- *prometheus.Desc) {
	descs <- l.price
}

func (l price) Collect(metrics chan<- prometheus.Metric) {
	lmp, err := l.client.LMP(context.Background())
	if err != nil {
		metrics <- prometheus.NewInvalidMetric(l.price, err)
		return
	}

	for kind, data := range map[string]miso.LMPTime{
		"1h":     lmp.HourlyIntegratedLMP,
		"5min":   lmp.FiveMinuteLMP,
		"exante": lmp.DayAheadExAnteLMP,
		"expost": lmp.DayAheadExPostLMP,
	} {
		for _, node := range data.Nodes {
			emit(metrics, prometheus.MustNewConstMetric(l.price, prometheus.GaugeValue, float64(node.LMP), kind, node.Region, node.Name, "LMP"), data.StartAt, data.EndAt)
			emit(metrics, prometheus.MustNewConstMetric(l.price, prometheus.GaugeValue, float64(node.MLC), kind, node.Region, node.Name, "MLC"), data.StartAt, data.EndAt)
			emit(metrics, prometheus.MustNewConstMetric(l.price, prometheus.GaugeValue, float64(node.MCC), kind, node.Region, node.Name, "MCC"), data.StartAt, data.EndAt)
		}
	}
}

func NewPrice(client *miso.Client) prometheus.Collector {
	return &price{
		client: client,

		price: prometheus.NewDesc("miso_price_usd",
			"The price for power at a certain place and time",
			[]string{"kind", "region", "node", "comp"}, nil,
		),
	}
}
