package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

type instrumentedTransport struct {
	http.RoundTripper
	counter *prometheus.CounterVec
}

func (c *instrumentedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := c.RoundTripper.RoundTrip(req)
	statusCode := "unknown"
	if resp != nil {
		statusCode = resp.Status
	}
	c.counter.With(prometheus.Labels{
		"host":   req.Host,
		"method": req.Method,
		"status": statusCode,
	}).Inc()

	return resp, err
}

func instrumentTransport(
	counter *prometheus.CounterVec,
	transport http.RoundTripper,
) http.RoundTripper {
	return &instrumentedTransport{
		RoundTripper: transport,
		counter:      counter,
	}
}
