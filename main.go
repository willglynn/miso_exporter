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

func main() {
	opts := promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}
	mux := http.NewServeMux()

	c := miso.NewClient(http.DefaultTransport)

	metrics := prometheus.NewRegistry()
	metrics.MustRegister(exporter.NewLoad(c))
	metrics.MustRegister(exporter.NewPrice(c))
	mux.Handle("/metrics", promhttp.HandlerFor(metrics, opts))

	realtime := prometheus.NewRegistry()
	realtime.MustRegister(exporter.NewLoad(c))
	mux.Handle("/load", promhttp.HandlerFor(realtime, opts))

	lmp := prometheus.NewRegistry()
	lmp.MustRegister(exporter.NewPrice(c))
	mux.Handle("/lmp", promhttp.HandlerFor(lmp, opts))

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
