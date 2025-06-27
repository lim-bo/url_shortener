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
	HTTPRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of http requests",
	}, []string{"method", "path"})
	HTTPDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_requests_duration",
		Help:    "http request proccessing duration",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"path"})
	Reg = prometheus.NewRegistry()
)

func init() {
	Reg.MustRegister(collectors.NewGoCollector())
	Reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	Reg.MustRegister(HTTPRequests)
	Reg.MustRegister(HTTPDuration)
	go func() {
		r := chi.NewMux()
		r.Handle("/metrics", promhttp.HandlerFor(Reg, promhttp.HandlerOpts{
			Registry: Reg,
		}))
		log.Fatal(http.ListenAndServe(":2112", r))
	}()
}
