package internal

import (
	"context"
	"github.com/felixge/httpsnoop"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"net/http"
)

type REDMetric struct {
	histogram instrument.Float64Histogram
}

func NewREDMetric(histogram instrument.Float64Histogram) *REDMetric {
	return &REDMetric{histogram}
}

func (metric *REDMetric) Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			metrics := httpsnoop.CaptureMetrics(next, writer, request)
			attrs := []attribute.KeyValue{
				attribute.Key("method").String(request.Method),
				attribute.Key("route").String(chi.RouteContext(request.Context()).RoutePattern()),
				attribute.Key("status_code").Int(metrics.Code),
			}
			metric.histogram.Record(context.Background(), metrics.Duration.Seconds(), attrs...)
		})
	}
}
