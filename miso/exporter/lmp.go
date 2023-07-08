package exporter

import (
	"context"
	"time"

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
			lmp := prometheus.MustNewConstMetric(l.price, prometheus.GaugeValue, float64(node.LMP), kind, node.Region, node.Name, "LMP")
			mlc := prometheus.MustNewConstMetric(l.price, prometheus.GaugeValue, float64(node.MLC), kind, node.Region, node.Name, "MLC")
			mcc := prometheus.MustNewConstMetric(l.price, prometheus.GaugeValue, float64(node.MCC), kind, node.Region, node.Name, "MCC")
			for t := data.StartAt; t.Before(data.EndAt); t = t.Add(time.Minute) {
				metrics <- prometheus.NewMetricWithTimestamp(t, lmp)
				metrics <- prometheus.NewMetricWithTimestamp(t, mlc)
				metrics <- prometheus.NewMetricWithTimestamp(t, mcc)
			}
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
