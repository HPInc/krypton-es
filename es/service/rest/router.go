// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/HPInc/krypton-es/es/service/metrics"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func requestLogger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Extract the request ID if specified, else create a new request ID.
		if r.Header.Get(headerRequestID) == "" {
			r.Header.Set(headerRequestID, uuid.NewString())
		}

		// Calculate and report REST latency metric.
		defer metrics.ReportLatencyMetric(metrics.MetricRestLatency, start,
			r.Method)

		if debugLogRestRequests {
			dump, err := httputil.DumpRequest(r, true)
			if err != nil {
				esLogger.Error("Error logging request!",
					zap.Error(err),
				)
				return
			}
			esLogger.Debug("+++ New REST request +++",
				zap.ByteString("Request", dump),
			)
		}

		inner.ServeHTTP(w, r)
		metrics.MetricRequestCount.Inc()

		if debugLogRestRequests {
			esLogger.Debug("-- Served REST request --",
				zap.String("Method: ", r.Method),
				zap.String("Request URI: ", r.RequestURI),
				zap.String("Route name: ", name),
				zap.String("Duration: ", time.Since(start).String()),
			)
		}
	})
}

// Initializes the REST request router for the FS service and registers all
// routes and their corresponding handler functions.
func initRequestRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.MethodNotAllowedHandler = http.HandlerFunc(esMethodNotAllowed)

	for _, route := range registeredRoutes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = requestLogger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Path).
			Name(route.Name).
			Handler(handler)
	}
	return router
}

// methodNotAllowed replies to the request with an HTTP status code 405.
// per RFC2616, 405 responses should have an Allow header
// Eg: if POST was expected, Allow: POST
func esMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	currentMethod := r.Method
	allMethods := []string{"GET", "DELETE", "PATCH", "PUT", "POST"}
	var allowMethods []string

	for _, m := range allMethods {
		if m == currentMethod {
			continue
		}
		var match mux.RouteMatch
		r.Method = m
		if router.Match(r, &match) && match.Route != nil {
			allowMethods = append(allowMethods, m)
		}
	}
	w.Header().Set("Allow", strings.Join(allowMethods[:], ","))
	w.WriteHeader(http.StatusMethodNotAllowed)
}
