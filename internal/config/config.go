// Copyright 2025 William Marchesi

// Author: William Marchesi
// Email: will@marchesi.io
// Website: https://marchesi.io/

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"os"
)

// Config holds the application configuration
// This reuses the same environment variables as spool-scanner
type Config struct {
	SpoolmanURL string    `json:"spoolman_url"`
	Printers    []Printer `json:"printers"`
}

// Printer configuration matching spool-scanner
type Printer struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	OctoPrintURL string `json:"octoprint_url"`
	APIKey       string `json:"api_key"`
}

// LoadConfig loads configuration from environment variables
// Uses same env vars as spool-scanner for consistency
func LoadConfig() (*Config, error) {
	cfg := &Config{
		SpoolmanURL: os.Getenv("SPOOLMAN_URL"),
		Printers:    []Printer{},
	}

	// Load up to 10 printers (same as spool-scanner)
	for i := 1; i <= 10; i++ {
		name := os.Getenv(fmt.Sprintf("PRINTER_%d_NAME", i))
		if name == "" {
			continue
		}

		printer := Printer{
			ID:           fmt.Sprintf("printer-%d", i),
			Name:         name,
			OctoPrintURL: os.Getenv(fmt.Sprintf("PRINTER_%d_URL", i)),
			APIKey:       os.Getenv(fmt.Sprintf("PRINTER_%d_KEY", i)),
		}

		if printer.OctoPrintURL == "" || printer.APIKey == "" {
			return nil, fmt.Errorf("incomplete configuration for printer %d", i)
		}

		cfg.Printers = append(cfg.Printers, printer)
	}

	if len(cfg.Printers) == 0 {
		return nil, fmt.Errorf("no printers configured")
	}

	if cfg.SpoolmanURL == "" {
		return nil, fmt.Errorf("SPOOLMAN_URL not set")
	}

	return cfg, nil
}
