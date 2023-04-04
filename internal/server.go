package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
	Router     chi.Router
}

func NewServer(logger *zap.Logger, port int, middlewares ...func(http.Handler) http.Handler) *Server {
	histogram := createHistogram(logger)
	router := chi.NewRouter()

	router.Use(ZapLogger(logger))
	router.Use(NewREDMetric(histogram).Middleware())
	router.Use(middlewares...)
	router.Handle("/metrics", promhttp.Handler())

	return &Server{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: router,
		},
		logger: logger,
		Router: router,
	}
}

func createHistogram(logger *zap.Logger) instrument.Float64Histogram {
	exporter, err := prometheus.New()
	if err != nil {
		logger.Fatal("couldnt create new exporter", zap.Error(err))
	}
	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	meter := provider.Meter("")

	histogram, err := meter.Float64Histogram("request_duration_seconds", instrument.WithDescription("Time (in seconds) spent serving HTTP requests."))
	if err != nil {
		logger.Fatal("couldnt create new histogram", zap.Error(err))
	}
	return histogram
}

// StartAndWait starts the http server and waits for sigint || sigterm. If it receives a signal it gracefully shutdowns the server
func (server *Server) StartAndWait() {
	logger := server.logger
	logger.Info("starting server", zap.String("addr", server.httpServer.Addr))

	go func() {
		err := server.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("HTTP server error", zap.Error(err))
		}
		logger.Info("Stopped serving new connections")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sigString := (<-sigChan).String()
	logger.Info("Shutdown Signal", zap.String("sig", sigString))

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("HTTP shutdown error", zap.Error(err))
	}
	logger.Info("Graceful shutdown complete")
}
