package internal

import (
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func ZapLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			ww := middleware.NewWrapResponseWriter(writer, request.ProtoMajor)
			timeNow := time.Now()
			defer func() {
				logger.Info("Served",
					zap.String("proto", request.Proto),
					zap.String("method", request.Method),
					zap.String("path", request.URL.Path),
					zap.Duration("lat", time.Since(timeNow)),
					zap.Int("status", ww.Status()),
					zap.Int("size", ww.BytesWritten()))
			}()
			next.ServeHTTP(ww, request)
		})
	}
}
