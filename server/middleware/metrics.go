package middleware

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/dzherb/mifi-go-microservice/metric"
)

func CollectRequestsMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metric.ActiveRequests.Inc()
		start := time.Now()

		rw := &interceptStatusResponseWriter{ResponseWriter: w}

		defer func() {
			metric.ActiveRequests.Dec()

			duration := time.Since(start).Seconds()
			statusCode := strconv.Itoa(rw.WrittenStatus())

			metric.TotalRequests.WithLabelValues(r.Method, r.URL.Path, statusCode).
				Inc()
			metric.RequestDuration.WithLabelValues(r.Method, r.URL.Path).
				Observe(duration)

			if rw.WrittenStatus() >= http.StatusBadRequest ||
				rw.WrittenStatus() == 0 {
				metric.ErrorsTotal.WithLabelValues(r.Method, r.URL.Path, statusCode).
					Inc()
			}
		}()

		next.ServeHTTP(rw, r)
	})
}

type interceptStatusResponseWriter struct {
	http.ResponseWriter
	statusCode atomic.Int64
}

func (rw *interceptStatusResponseWriter) WriteHeader(code int) {
	rw.statusCode.Store(int64(code))
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *interceptStatusResponseWriter) Write(b []byte) (int, error) {
	rw.statusCode.CompareAndSwap(0, http.StatusOK)

	return rw.ResponseWriter.Write(b)
}

func (rw *interceptStatusResponseWriter) WrittenStatus() int {
	return int(rw.statusCode.Load())
}
