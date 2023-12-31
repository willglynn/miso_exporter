package exporter

import (
	"context"

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

	for kind, datapoints := range map[string][]miso.RenewableDatapoint{
		"actual":   production.Actual,
		"forecast": production.Forecast,
	} {
		for _, entry := range datapoints {
			m := prometheus.MustNewConstMetric(c.renewableProduction, prometheus.GaugeValue, float64(entry.Megawatts)*1_000_000, kind)
			emit(metrics, m, entry.StartAt, entry.EndAt)
		}
	}
}

func NewRenewableProduction(client *miso.Client, source miso.Renewable) prometheus.Collector {
	return &renewableProduction{
		client: client,
		source: source,

		renewableProduction: prometheus.NewDesc("miso_renewable_production_w",
			"The amount of power produced from renewable sources",
			[]string{"kind"}, map[string]string{
				"source": source.String(),
			},
		),
	}
}
