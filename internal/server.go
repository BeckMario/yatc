package internal

import (
	"context"
	"errors"
	"fmt"
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
}

func NewServer(logger *zap.Logger, port int, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: handler,
		},
		logger: logger,
	}
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
