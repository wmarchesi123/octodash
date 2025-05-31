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

package spoolman

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client handles communication with Spoolman API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Spoolman client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Spool represents a filament spool
type Spool struct {
	ID              int     `json:"id"`
	Registered      string  `json:"registered"`
	FirstUsed       *string `json:"first_used"`
	LastUsed        *string `json:"last_used"`
	Price           float64 `json:"price"`
	RemainingWeight float64 `json:"remaining_weight"`
	InitialWeight   float64 `json:"initial_weight"`
	SpoolWeight     float64 `json:"spool_weight"`
	UsedWeight      float64 `json:"used_weight"`
	RemainingLength float64 `json:"remaining_length"`
	UsedLength      float64 `json:"used_length"`
	Comment         string  `json:"comment"`
	Location        string  `json:"location"`
	LotNr           string  `json:"lot_nr"`
	Archived        bool    `json:"archived"`
	Filament        struct {
		ID                   int     `json:"id"`
		Registered           string  `json:"registered"`
		Name                 string  `json:"name"`
		Material             string  `json:"material"`
		Price                float64 `json:"price"`
		Density              float64 `json:"density"`
		Diameter             float64 `json:"diameter"`
		Weight               float64 `json:"weight"`
		SpoolWeight          float64 `json:"spool_weight"`
		ArticleNr            string  `json:"article_number"`
		Comment              string  `json:"comment"`
		SettingsExtruderTemp float64 `json:"settings_extruder_temp"`
		SettingsBedTemp      float64 `json:"settings_bed_temp"`
		ColorHex             string  `json:"color_hex"`
		MultiColorHexes      string  `json:"multi_color_hexes"`
		MultiColorDirection  string  `json:"multi_color_direction"`
		ExternalID           string  `json:"external_id"`
		Vendor               struct {
			ID               int     `json:"id"`
			Registered       string  `json:"registered"`
			Name             string  `json:"name"`
			Comment          string  `json:"comment"`
			EmptySpoolWeight float64 `json:"empty_spool_weight"`
			ExternalID       string  `json:"external_id"`
		} `json:"vendor"`
	} `json:"filament"`
}

// GetSpool fetches a specific spool by ID
func (c *Client) GetSpool(id string) (*Spool, error) {
	url := fmt.Sprintf("%s/api/v1/spool/%s", c.baseURL, id)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("spool not found")
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var spool Spool
	if err := json.NewDecoder(resp.Body).Decode(&spool); err != nil {
		return nil, err
	}

	return &spool, nil
}

// FormatSpoolInfo creates a formatted spool info structure
func FormatSpoolInfo(spool *Spool) map[string]interface{} {
	if spool == nil {
		return nil
	}

	// Format color
	color := "#888888"
	if spool.Filament.ColorHex != "" {
		color = spool.Filament.ColorHex
		if color[0] != '#' {
			color = "#" + color
		}
	}

	// Build display name
	displayName := spool.Filament.Name
	if displayName == "" && spool.Filament.Material != "" {
		displayName = fmt.Sprintf("%s - %s", spool.Filament.Vendor.Name, spool.Filament.Material)
	}

	return map[string]interface{}{
		"id":        fmt.Sprintf("%d", spool.ID),
		"name":      displayName,
		"material":  spool.Filament.Material,
		"color":     color,
		"vendor":    spool.Filament.Vendor.Name,
		"weight":    spool.InitialWeight,
		"used":      spool.UsedWeight,
		"remaining": spool.RemainingWeight,
	}
}
