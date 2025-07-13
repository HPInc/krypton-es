// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HPInc/krypton-es/es/service/config"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const (
	apiUrlPrefix      = "/api/v1"
	apiInternalPrefix = "/api/v1/internal"

	// UUIDRegex defines a pattern for validating UUIDs
	uuidRegex = "[a-fA-F0-9]{8}-?[a-fA-F0-9]{4}-?4[a-fA-F0-9]{3}-?[8|9|aA|bB][a-fA-F0-9]{3}-?[a-fA-F0-9]{12}"
)

var (
	// Structured logging using Uber Zap.
	esLogger *zap.Logger

	debugLogRestRequests bool

	// Connection to the devices database.
	gSrv *http.Server

	// Error channel
	errorChannel chan error

	// Interrupt channel
	interruptChannel chan os.Signal

	// router
	router *mux.Router

	// server config
	gServerConfig *config.Server
)

func Init(logger *zap.Logger, serverConfig *config.Server) error {
	esLogger = logger
	debugLogRestRequests = serverConfig.DebugRestRequests
	gServerConfig = serverConfig

	errorChannel = make(chan error)
	interruptChannel = make(chan os.Signal, 1)
	signal.Notify(interruptChannel, syscall.SIGINT, syscall.SIGTERM)

	router = initRequestRouter()
	addr := fmt.Sprintf("%s:%d", serverConfig.Host, serverConfig.Port)
	gSrv = &http.Server{
		Addr: addr,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}
	esLogger.Info("Starting Enrollment REST server: ",
		zap.String("Address:", addr),
	)

	go func() {
		if err := gSrv.ListenAndServe(); err != nil {
			esLogger.Error("Server error", zap.Error(err))
			errorChannel <- err
		}
	}()
	return nil
}

func Shutdown() {
	var err error
	if gSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		if err = gSrv.Shutdown(ctx); err != nil {
			esLogger.Error("There was an error shutting down server", zap.Error(err))
		} else {
			esLogger.Info("Server was shutdown successfully.")
		}
	} else {
		esLogger.Info("Not shutting down server as it was not initialized")
	}
}

// Wait for a signal to shutdown the web server and cleanup.
func WaitForEvents() {
	// Block until we receive either an OS signal, or encounter a server
	// fatal error and need to terminate.
	select {
	case err := <-errorChannel:
		esLogger.Error("Device Enrollment Service: Shutting down due to a fatal error.",
			zap.Error(err),
		)
	case sig := <-interruptChannel:
		esLogger.Info("Device Enrollment Service: Received an OS signal and shutting down.",
			zap.String("Signal:", sig.String()),
		)
	}
}
