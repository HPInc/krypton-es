// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	esLogger *zap.Logger
	logLevel zap.AtomicLevel
)

func InitTestLogger() {
	_ = os.Setenv("CONSOLE_LOGGING", "enabled")
	initLogger("debug")
}

func initLogger(levelString string) {
	// Log to the console by default.
	logLevel = zap.NewAtomicLevel()
	encoderCfg := zap.NewProductionEncoderConfig()
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		logLevel)
	esLogger = zap.New(core, zap.AddCaller())
	setLogLevel(levelString)
}

func shutdownLogger() {
	_ = esLogger.Sync()
}

func setLogLevel(level string) {
	parsedLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		// Fallback to logging at the info level.
		fmt.Printf("Falling back to the info log level. You specified: %s.\n",
			level)
		logLevel.SetLevel(zapcore.InfoLevel)
	} else {
		logLevel.SetLevel(parsedLevel)
	}
}
