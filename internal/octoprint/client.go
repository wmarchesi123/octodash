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

package octoprint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client handles communication with OctoPrint API
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new OctoPrint client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// PrinterState represents the current printer state
type PrinterState struct {
	Text  string `json:"text"`
	Flags struct {
		Operational bool `json:"operational"`
		Paused      bool `json:"paused"`
		Printing    bool `json:"printing"`
		Error       bool `json:"error"`
		Ready       bool `json:"ready"`
	} `json:"flags"`
}

// TemperatureData represents temperature information
type TemperatureData struct {
	Actual float64 `json:"actual"`
	Target float64 `json:"target"`
}

// PrinterResponse represents the full printer API response
type PrinterResponse struct {
	State       PrinterState `json:"state"`
	Temperature struct {
		Bed   TemperatureData `json:"bed"`
		Tool0 TemperatureData `json:"tool0"`
	} `json:"temperature"`
}

// JobResponse represents print job information
type JobResponse struct {
	Job struct {
		File struct {
			Name    string `json:"name"`
			Size    int64  `json:"size"`
			Date    int64  `json:"date"`
			Path    string `json:"path"`
			Display string `json:"display"`
		} `json:"file"`
		EstimatedPrintTime float64 `json:"estimatedPrintTime"`
		LastPrintTime      float64 `json:"lastPrintTime"`
		Filament           struct {
			Tool0 struct {
				Length float64 `json:"length"`
				Volume float64 `json:"volume"`
			} `json:"tool0"`
		} `json:"filament"`
	} `json:"job"`
	Progress struct {
		Completion      float64 `json:"completion"`
		Filepos         int64   `json:"filepos"`
		PrintTime       int     `json:"printTime"`
		PrintTimeLeft   int     `json:"printTimeLeft"`
		PrintTimeOrigin string  `json:"printTimeOrigin"`
	} `json:"progress"`
	State string `json:"state"`
}

// GetPrinterState fetches current printer state and temperatures
func (c *Client) GetPrinterState() (*PrinterResponse, error) {
	req, err := c.newRequest("GET", "/api/printer", nil)
	if err != nil {
		return nil, err
	}

	var response PrinterResponse
	if err := c.doRequest(req, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetJob fetches current job information
func (c *Client) GetJob() (*JobResponse, error) {
	req, err := c.newRequest("GET", "/api/job", nil)
	if err != nil {
		return nil, err
	}

	var response JobResponse
	if err := c.doRequest(req, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetCurrentSpool fetches the currently selected spool ID
func (c *Client) GetCurrentSpool(tool int) (string, error) {
	payload := map[string]interface{}{
		"command": "get_current_spool",
		"tool":    tool,
	}

	req, err := c.newRequest("POST", "/api/plugin/spoolman_api", payload)
	if err != nil {
		return "", err
	}

	var response struct {
		Success bool   `json:"success"`
		SpoolID string `json:"spool_id"`
		Error   string `json:"error,omitempty"`
	}

	if err := c.doRequest(req, &response); err != nil {
		return "", err
	}

	if !response.Success {
		return "", fmt.Errorf("API error: %s", response.Error)
	}

	return response.SpoolID, nil
}

// GetThumbnail fetches the thumbnail URL for the current job
func (c *Client) GetThumbnail(path string) string {
	// OctoPrint stores thumbnails at a predictable path
	// This returns the URL - the actual fetching happens in the browser
	if path == "" {
		return ""
	}

	// Remove .gcode extension if present and add .png
	fileName := path

	fileName = strings.TrimSuffix(fileName, ".gcode")
	fileName = strings.TrimSuffix(fileName, ".bgcode")

	return fmt.Sprintf("%s/plugin/prusaslicerthumbnails/thumbnail/%s.png", c.baseURL, fileName)
}

// Helper methods

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) doRequest(req *http.Request, result interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}

	return nil
}
