package main

import (
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/willglynn/miso_exporter/miso"
	"github.com/willglynn/miso_exporter/miso/exporter"
)

func handlerFor(collector ...prometheus.Collector) http.Handler {
	reg := prometheus.NewRegistry()
	for _, c := range collector {
		reg.MustRegister(c)
	}
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

func main() {
	mux := http.NewServeMux()

	c := miso.NewClient(http.DefaultTransport)

	load := exporter.NewLoad(c)
	lmp := exporter.NewPrice(c)

	solarProduction := exporter.NewRenewableProduction(c, miso.RenewableSolar)
	windProduction := exporter.NewRenewableProduction(c, miso.RenewableWind)

	mux.Handle("/metrics", handlerFor(load, lmp, solarProduction, windProduction))

	mux.Handle("/load", handlerFor(load))
	mux.Handle("/lmp", handlerFor(lmp))
	mux.Handle("/renewable_production", handlerFor(solarProduction, windProduction))

	var addr string
	addr = os.Getenv("LISTEN")
	if port := os.Getenv("PORT"); addr == "" && port != "" {
		addr = ":" + port
	}
	if addr == "" {
		addr = ":2023"
	}

	log.Printf("Starting HTTP server on %v", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
