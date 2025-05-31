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

package models

import "fmt"

// PrinterStatus represents the dashboard view of a printer
type PrinterStatus struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	OctoPrintURL string                 `json:"octoprint_url"`
	Status       string                 `json:"status"`
	State        string                 `json:"state"`
	Progress     *ProgressInfo          `json:"progress,omitempty"`
	Temperatures *TemperatureInfo       `json:"temperatures,omitempty"`
	CurrentSpool map[string]interface{} `json:"current_spool,omitempty"`
	ThumbnailURL string                 `json:"thumbnail_url,omitempty"`
	Error        string                 `json:"error,omitempty"`
}

// ProgressInfo represents print progress for the dashboard
type ProgressInfo struct {
	Completion     float64 `json:"completion"`
	PrintTime      int     `json:"print_time"`
	PrintTimeLeft  int     `json:"print_time_left"`
	EstimatedTotal int     `json:"estimated_total"`
	FileName       string  `json:"file_name"`
	FilamentLength float64 `json:"filament_length"`
}

// TemperatureInfo represents temperature data for the dashboard
type TemperatureInfo struct {
	BedActual    float64 `json:"bed_actual"`
	BedTarget    float64 `json:"bed_target"`
	HotendActual float64 `json:"hotend_actual"`
	HotendTarget float64 `json:"hotend_target"`
}

// FormatDuration formats seconds into a human-readable duration
func FormatDuration(seconds int) string {
	if seconds <= 0 {
		return "--:--:--"
	}

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}
