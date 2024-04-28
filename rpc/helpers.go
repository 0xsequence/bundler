package rpc

import (
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

func (s *RPC) renderJSON(w http.ResponseWriter, r *http.Request, v interface{}, status int) {
	buf, err := json.Marshal(v)
	if err != nil {
		s.GetLogger(r.Context()).Error("json.Marshal: failed to serialize response body", "err", err)
		// TODO: similar.. errorHandler(w, proto.ErrorInternal("failed to serialize response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(buf)
}

type metrics struct {
	methodTime *prometheus.HistogramVec
}

func createMetrics(reg prometheus.Registerer) *metrics {
	methodTime := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "rpc_method_time",
		Help:    "Method execution time",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 18),
	}, []string{"method"})

	if reg != nil {
		reg.MustRegister(methodTime)
	}

	return &metrics{
		methodTime: methodTime,
	}
}
