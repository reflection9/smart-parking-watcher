package observability

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type HTTPObserver struct {
	serviceName       string
	tracingMiddleware gin.HandlerFunc
	metricsMiddleware gin.HandlerFunc
	metricsHandler    http.Handler
	shutdown          func(context.Context) error
}

func NewHTTPObserver(ctx context.Context, serviceName, otlpEndpoint string) (*HTTPObserver, error) {
	observer := &HTTPObserver{
		serviceName:       serviceName,
		metricsMiddleware: newMetricsMiddleware(serviceName),
		metricsHandler:    promhttp.Handler(),
		shutdown:          func(context.Context) error { return nil },
	}

	if otlpEndpoint == "" {
		return observer, nil
	}

	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(otlpEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	observer.tracingMiddleware = otelgin.Middleware(serviceName)
	observer.shutdown = tracerProvider.Shutdown

	return observer, nil
}

func (o *HTTPObserver) Attach(router *gin.Engine) {
	router.Use(o.metricsMiddleware)
	if o.tracingMiddleware != nil {
		router.Use(o.tracingMiddleware)
	}
	router.GET("/metrics", gin.WrapH(o.metricsHandler))
}

func (o *HTTPObserver) Shutdown(ctx context.Context) error {
	return o.shutdown(ctx)
}

func NewHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}

func newMetricsMiddleware(serviceName string) gin.HandlerFunc {
	requestTotal := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "smart_parking_http_requests_total",
			Help: "Total number of handled HTTP requests.",
		},
		[]string{"service", "method", "route", "status"},
	)

	requestDuration := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "smart_parking_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "route", "status"},
	)

	inFlight := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "smart_parking_http_in_flight_requests",
			Help: "Current in-flight HTTP requests.",
		},
		[]string{"service"},
	)

	return func(c *gin.Context) {
		start := time.Now()
		inFlight.WithLabelValues(serviceName).Inc()
		defer inFlight.WithLabelValues(serviceName).Dec()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		requestTotal.WithLabelValues(serviceName, method, route, status).Inc()
		requestDuration.WithLabelValues(serviceName, method, route, status).Observe(time.Since(start).Seconds())
	}
}
