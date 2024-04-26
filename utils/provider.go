package utils

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type HttpRpcMetricsClient struct {
	transport http.RoundTripper

	// Metrics
	callTime   *prometheus.HistogramVec
	callErrors *prometheus.CounterVec
}

func NewHttpRpcMetricsClient() *HttpRpcMetricsClient {
	return &HttpRpcMetricsClient{
		transport: http.DefaultTransport,
		callTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "rpc_call_time",
			Help: "Time it takes to make an RPC call",
		}, []string{"method"}),
		callErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rpc_call_errors",
			Help: "Number of errors when making an RPC call",
		}, []string{"method"}),
	}
}

func tryObtainRpcMethod(req *http.Request) string {
	body, err := req.GetBody()
	if err != nil {
		return "unknown"
	}

	type rpcRequest struct {
		Method string `json:"method"`
	}

	var rpc rpcRequest
	err = json.NewDecoder(body).Decode(&rpc)
	if err != nil {
		return "unknown"
	}

	if rpc.Method != "" {
		return rpc.Method
	}

	return "unknown"
}

func (c *HttpRpcMetricsClient) RoundTrip(req *http.Request) (*http.Response, error) {
	method := tryObtainRpcMethod(req)

	start := time.Now()
	resp, err := c.transport.RoundTrip(req)
	duration := time.Since(start)

	c.callTime.WithLabelValues(method).Observe(float64(duration.Seconds()))

	if err != nil {
		c.callErrors.WithLabelValues(method).Inc()
	}

	return resp, err
}

func (c *HttpRpcMetricsClient) UseRegistry(reg prometheus.Registerer, tag string) {
	tagged := prometheus.WrapRegistererWith(prometheus.Labels{"tag": tag}, reg)
	tagged.MustRegister(
		c.callTime,
		c.callErrors,
	)
}
