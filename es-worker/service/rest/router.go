// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

// Purpose:
// Initializes the request router (Gorilla MUX) that is used to serve REST
// requests at ESW's REST endpoint. A common wrapped HTTP request handler
// is attached to various HTTP handler functions to perform common tasks like
// extracting or issuing request IDs, measuring latency metrics and optionally
// debug logging REST requests.
package rest

import (
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/HPInc/krypton-es/es-worker/service/metrics"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Common HTTP handler function invoked to handle all requests received at
// the REST server. It performs a few tasks centrally across all requests:
//   - Generating a request ID for requests that do not have one.
//   - Calculating latency metrics.
//   - Logging requests if configured to do so for debugging purposes.
func commonRequestHandler(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Extract the request ID if specified, else create a new request ID.
		if r.Header.Get(headerRequestID) == "" {
			r.Header.Set(headerRequestID, uuid.New().String())
		}

		// Calculate and report REST latency metric.
		defer metrics.ReportLatencyMetric(metrics.MetricRestLatency, start,
			r.Method)

		if debugLogRestRequests {
			dump, err := httputil.DumpRequest(r, true)
			if err != nil {
				eswLogger.Error("Error logging request!",
					zap.String("Method: ", r.Method),
					zap.String("Request URI: ", r.RequestURI),
					zap.String("Route name: ", name),
					zap.Error(err),
				)
				return
			}
			eswLogger.Debug("+++ New REST request +++",
				zap.ByteString("Request", dump),
			)
		}

		inner.ServeHTTP(w, r)
		if debugLogRestRequests {
			eswLogger.Debug("-- Served REST request --",
				zap.String("Method: ", r.Method),
				zap.String("Request URI: ", r.RequestURI),
				zap.String("Route name: ", name),
				zap.String("Duration: ", time.Since(start).String()),
			)
		}
	})
}

// Initializes the REST request router for the ESW service and registers all
// routes and their corresponding handler functions.
func initRequestRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	for _, route := range registeredRoutes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = commonRequestHandler(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Path).
			Name(route.Name).
			Handler(handler)
	}
	return router
}
