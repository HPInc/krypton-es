// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

// Purpose:
// Initialize HTTP REST server for ESW (enrollment worker) service.
package rest

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HPInc/krypton-es/es-worker/service/config"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	eswLogger            *zap.Logger
	debugLogRestRequests bool
)

const (
	// HTTP server timeouts for the REST endpoint.
	readTimeout  = (time.Second * 5)
	writeTimeout = (time.Second * 5)
)

// Represents the es worker REST service.
type eswRestService struct {
	// Signal handling to support SIGTERM and SIGINT for the service.
	errChannel  chan error
	stopChannel chan os.Signal

	// Prometheus metrics reporting.
	metricRegistry *prometheus.Registry

	// Request router
	router *mux.Router

	// HTTP port on which the REST server is available.
	port int
}

// Creates a new instance of ESW REST service
// initalizes request router for the ESW REST endpoint.
func NewService() *eswRestService {
	s := &eswRestService{}

	// Initial signal handling.
	s.errChannel = make(chan error)
	s.stopChannel = make(chan os.Signal, 1)
	signal.Notify(s.stopChannel, syscall.SIGINT, syscall.SIGTERM)

	// Initialize the prometheus metric reporting registry.
	s.metricRegistry = prometheus.NewRegistry()

	s.router = initRequestRouter()
	return s
}

// Starts the HTTP REST server for the ESW service and starts serving requests
// at the REST endpoint.
func (s *eswRestService) startServing() {
	// Start the HTTP REST server. http.ListenAndServe() always returns
	// a non-nil error
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", s.port),
		Handler:        s.router,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	err := server.ListenAndServe()
	eswLogger.Error("Received a fatal error from http.ListenAndServe",
		zap.Error(err),
	)

	// Signal the error channel so we ESWn shutdown the service.
	s.errChannel <- err
}

// Waits for ESW REST server to be terminated - either in response to a
// system event received on the stop channel or a fatal error signal received
// on the error channel.
func (s *eswRestService) awaitTermination() {
	select {
	case err := <-s.errChannel:
		eswLogger.Error("Shutting down due to a fatal error.",
			zap.Error(err),
		)
	case sig := <-s.stopChannel:
		eswLogger.Info("Received an OS signal to shut down!",
			zap.String("Signal received: ", sig.String()),
		)
	}
}

// Initializes ESW REST server and starts serving requests
func Init(logger *zap.Logger, serverConfig *config.Server) {
	eswLogger = logger
	debugLogRestRequests = serverConfig.DebugLogRestRequests

	s := NewService()
	s.port = serverConfig.Port

	// Initialize the REST server and listen for REST requests on a separate
	// goroutine. Report fatal errors via the error channel.
	go s.startServing()
	eswLogger.Info("Started the ESW REST service!",
		zap.Int("Port: ", s.port),
	)

	// Wait for the REST server to be terminated either in response to a system
	// event (like service shutdown) or a fatal error.
	s.awaitTermination()
}
