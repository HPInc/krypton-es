// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package tokenmgr

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type TokenType string

// token attributes
type TokenIssuerSettings struct {
	Type            string `yaml:"type"`
	KeysURL         string `yaml:"keys"`
	Audience        string `yaml:"audience"`
	Issuer          string `yaml:"issuer"`
	RefreshInterval int    `yaml:"refresh_interval"`
	DefaultTenantId string `yaml:"default_tenant_id"`
	// app token auth details
	AllowedAppIds []string `yaml:"allowed_app_ids"`
}

type Config struct {
	TokenTypes map[TokenType]TokenIssuerSettings `yaml:"token_types"`
}

var tokenConfig Config

func loadTokenConfiguration(tokenConfigFile string) bool {
	// Open the configuration file for parsing.
	bytes, err := os.ReadFile(filepath.Clean(tokenConfigFile))
	if err != nil {
		esLogger.Error("Failed to load configuration file!",
			zap.String("Configuration file:", tokenConfigFile),
			zap.Error(err),
		)
		return false
	}

	// Read the configuration file and unmarshal the YAML.
	err = yaml.Unmarshal(bytes, &tokenConfig)
	if err != nil {
		esLogger.Error("Failed to parse configuration file!",
			zap.String("Configuration file:", tokenConfigFile),
			zap.Error(err),
		)
		return false
	}

	esLogger.Info("Parsed configuration from the token configuration file!",
		zap.String("Configuration file:", tokenConfigFile),
	)
	return true
}
