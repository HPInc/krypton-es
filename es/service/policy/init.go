// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package policy

import (
	"encoding/json"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

const (
	defaultPolicyVersion = 1
)

var (
	// Structured logging using Uber Zap.
	esLogger      *zap.Logger
	defaultPolicy *Policy
)

func Init(logger *zap.Logger, policyFile string) error {
	var err error
	esLogger = logger

	if defaultPolicy, err = loadPolicy(policyFile); err != nil {
		return ErrLoadDefaultPolicy
	}
	esLogger.Info("Default policy", zap.Any("data", defaultPolicy))
	return nil
}

func loadPolicy(policyFile string) (*Policy, error) {
	// Open the policy file for parsing.
	bytes, err := os.ReadFile(filepath.Clean(policyFile))
	if err != nil {
		esLogger.Error("Failed to load policy file!",
			zap.String("Policy file:", policyFile),
			zap.Error(err),
		)
		return nil, err
	}

	data := Policy{}
	// Read the config file and unmarshal json.
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		esLogger.Error("Failed to parse policy file!",
			zap.String("Policy file:", policyFile),
			zap.Error(err),
		)
		return nil, err
	}

	if err = data.Validate(); err != nil {
		return nil, err
	}

	esLogger.Info("Parsed policy data from file!",
		zap.String("File:", policyFile),
	)
	return &data, nil
}

func GetDefault() *Policy {
	return defaultPolicy
}
