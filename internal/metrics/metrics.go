package metrics

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of http requests",
	}, []string{"method", "path"})
	httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_requests_duration",
		Help:    "http request proccessing duration",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"path"})
)

func RegisterMetrics() {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	reg.MustRegister(httpRequests)
	reg.MustRegister(httpDuration)
	go func() {
		r := chi.NewMux()
		r.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
		log.Fatal(http.ListenAndServe(":2112", r))
	}()
}
