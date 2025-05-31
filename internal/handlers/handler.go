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

package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/wmarchesi123/octodash/internal/config"
	"github.com/wmarchesi123/octodash/internal/models"
	"github.com/wmarchesi123/octodash/internal/octoprint"
	"github.com/wmarchesi123/octodash/internal/spoolman"
)

// Handler manages HTTP routes and dependencies
type Handler struct {
	config           *config.Config
	mux              *http.ServeMux
	octoprintClients map[string]*octoprint.Client
	spoolmanClient   *spoolman.Client
}

// NewHandler creates a new handler with all routes configured
func NewHandler() *Handler {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	h := &Handler{
		config:           cfg,
		mux:              http.NewServeMux(),
		octoprintClients: make(map[string]*octoprint.Client),
		spoolmanClient:   spoolman.NewClient(cfg.SpoolmanURL),
	}

	// Create OctoPrint clients for each printer
	for _, printer := range cfg.Printers {
		h.octoprintClients[printer.ID] = octoprint.NewClient(printer.OctoPrintURL, printer.APIKey)
	}

	// Set up all routes
	h.setupRoutes()

	return h
}

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers for API calls
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	h.mux.ServeHTTP(w, r)
}

// setupRoutes configures all HTTP routes
func (h *Handler) setupRoutes() {
	// Static files (CSS, JS, images)
	h.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Main dashboard
	h.mux.HandleFunc("/", h.handleDashboard)

	// API endpoint for printer status (will poll every second)
	h.mux.HandleFunc("/api/status", h.handleStatus)
}

// handleDashboard serves the main dashboard HTML
func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	tmplStr := `
<!DOCTYPE html>
<html>
<head>
    <title>OctoDash - Printer Dashboard</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/style.css">
    <script src="//unpkg.com/alpinejs" defer></script>
</head>
<body>
    <div class="dashboard" x-data="dashboard" x-init="init()">
        <!-- Loading State -->
        <div x-show="loading" class="loading-overlay">
            <div class="loading-spinner"></div>
            <p>Loading printer status...</p>
        </div>

        <!-- Error State -->
        <div x-show="error" class="error-message">
            <p x-text="error"></p>
        </div>

        <!-- Printer Grid -->
        <div x-show="!loading && !error" class="printer-grid" :class="'printers-' + printers.length">
            <template x-for="printer in printers" :key="printer.id">
                <div class="printer-card" @click="openPrinter(printer)">
                    <h2 class="printer-name" x-text="printer.name"></h2>
                    
                    <!-- Printer Image Area -->
                    <div class="printer-image">
                        <!-- Show thumbnail if printing, otherwise stock image -->
                        <img :src="printer.thumbnail_url || '/static/prusa-mk4s.png'" 
                             :alt="printer.name"
                             @error="$event.target.src = '/static/prusa-mk4s.png'">
                    </div>

                    <!-- Status Area -->
                    <div class="printer-status" :class="'status-' + printer.status">
                        <div class="status-text">
                            <span class="status-label">Status:</span>
                            <span class="status-value" x-text="formatStatus(printer.status)"></span>
                        </div>

                        <!-- Progress Bar (if printing) -->
                        <div x-show="printer.progress" class="progress-section">
                            <div class="progress-bar">
                                <div class="progress-fill" :style="'width: ' + (printer.progress?.completion || 0) + '%'"></div>
                            </div>
                            <div class="progress-text">
                                <span x-text="Math.round(printer.progress?.completion || 0) + '%'"></span>
                            </div>
                        </div>

						<!-- Current Spool Info -->
						<div x-show="printer.current_spool" class="spool-info">
							<div class="spool-header">
								<span class="spool-color-dot" 
									:style="'background-color: ' + (printer.current_spool?.color || '#888')"></span>
								<div class="spool-title">
									<div class="spool-name">
										<span x-text="printer.current_spool?.name || 'Unknown'"></span>
										<span class="spool-material" x-text="' | ' + (printer.current_spool?.material || '')"></span>
									</div>
									<div class="spool-vendor" x-text="printer.current_spool?.vendor"></div>
									<div class="spool-stats">
										<span class="stat-item">
											<span class="stat-label">Total Weight</span>
											<span class="stat-value" x-text="formatWeight(printer.current_spool?.weight)"></span>
										</span>
										<span class="stat-separator"> | </span>
										<span class="stat-item">
											<span class="stat-label">Used</span>
											<span class="stat-value" x-text="formatWeight(printer.current_spool?.used)"></span>
										</span>
										<span class="stat-separator"> | </span>
										<span class="stat-item">
											<span class="stat-label">Remaining</span>
											<span class="stat-value" x-text="formatWeight(printer.current_spool?.remaining)"></span>
										</span>
									</div>
								</div>
								<div class="spool-progress">
									<svg class="progress-ring" width="80" height="80">
										<circle class="progress-ring-bg" cx="40" cy="40" r="35" />
										<circle class="progress-ring-fill" 
												cx="40" cy="40" r="35"
												:stroke="printer.current_spool?.color || '#888'"
												:stroke-dasharray="2 * Math.PI * 35"
												:stroke-dashoffset="2 * Math.PI * 35 * (1 - (printer.current_spool?.remaining || 0) / (printer.current_spool?.weight || 1))" />
									</svg>
									<div class="progress-text" x-text="Math.round(((printer.current_spool?.remaining || 0) / (printer.current_spool?.weight || 1)) * 100) + '%'"></div>
								</div>
							</div>
						</div>

                        <!-- Print Time Info -->
                        <div x-show="printer.progress" class="time-info">
                            <div class="time-item">
                                <span class="time-label">Elapsed:</span>
                                <span x-text="formatTime(printer.progress?.print_time)"></span>
                            </div>
                            <div class="time-item">
                                <span class="time-label">Remaining:</span>
                                <span x-text="formatTime(printer.progress?.print_time_left)"></span>
                            </div>
                        </div>

                        <!-- Temperature Info -->
                        <div x-show="printer.temperatures" class="temp-info">
                            <div class="temp-item">
                                <span class="temp-label">Hotend:</span>
                                <span x-text="formatTemp(printer.temperatures?.hotend_actual, printer.temperatures?.hotend_target)"></span>
                            </div>
                            <div class="temp-item">
                                <span class="temp-label">Bed:</span>
                                <span x-text="formatTemp(printer.temperatures?.bed_actual, printer.temperatures?.bed_target)"></span>
                            </div>
                        </div>
                    </div>
                </div>
            </template>
        </div>

        <!-- Return Overlay (hidden by default) -->
        <div x-show="showReturnOverlay" class="return-overlay" style="display: none;">
            <button @click="returnToDashboard()" class="return-button">
                ← Return to Dashboard
            </button>
        </div>
    </div>

    <script>
        // Configuration passed from server
        const PRINTERS = {{.PrintersJSON}};
    </script>
    <script src="/static/app.js"></script>
</body>
</html>
`

	// Prepare printer configuration for frontend
	printers := make([]map[string]string, len(h.config.Printers))
	for i, p := range h.config.Printers {
		printers[i] = map[string]string{
			"id":            p.ID,
			"name":          p.Name,
			"octoprint_url": p.OctoPrintURL,
		}
	}

	printersJSON, _ := json.Marshal(printers)

	tmpl, err := template.New("dashboard").Parse(tmplStr)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := struct {
		PrintersJSON template.JS
	}{
		PrintersJSON: template.JS(printersJSON),
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, data)
}

// handleStatus returns current printer status as JSON
func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	// Fetch status for all printers in parallel
	var wg sync.WaitGroup
	statusChan := make(chan *models.PrinterStatus, len(h.config.Printers))

	for _, printer := range h.config.Printers {
		wg.Add(1)
		go func(p config.Printer) {
			defer wg.Done()
			status := h.fetchPrinterStatus(p)
			statusChan <- status
		}(printer)
	}

	// Wait for all fetches to complete
	wg.Wait()
	close(statusChan)

	// Collect results
	printers := make([]*models.PrinterStatus, 0, len(h.config.Printers))
	for status := range statusChan {
		printers = append(printers, status)
	}

	// Sort by printer ID to maintain consistent order
	// (In a real implementation, you might want to sort by printer.ID)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "ok",
		"printers": printers,
	})
}

// fetchPrinterStatus fetches status for a single printer
func (h *Handler) fetchPrinterStatus(printer config.Printer) *models.PrinterStatus {
	status := &models.PrinterStatus{
		ID:           printer.ID,
		Name:         printer.Name,
		OctoPrintURL: printer.OctoPrintURL,
		Status:       "offline",
	}

	client, ok := h.octoprintClients[printer.ID]
	if !ok {
		status.Error = "No client configured"
		return status
	}

	// Fetch printer state and temperatures
	printerResp, err := client.GetPrinterState()
	if err != nil {
		log.Printf("Error fetching printer state for %s: %v", printer.Name, err)
		status.Error = err.Error()
		return status
	}

	// Set basic status
	if printerResp.State.Flags.Printing {
		status.Status = "printing"
	} else if printerResp.State.Flags.Ready {
		status.Status = "idle"
	} else if printerResp.State.Flags.Error {
		status.Status = "error"
	}
	status.State = printerResp.State.Text

	// Set temperatures
	status.Temperatures = &models.TemperatureInfo{
		BedActual:    printerResp.Temperature.Bed.Actual,
		BedTarget:    printerResp.Temperature.Bed.Target,
		HotendActual: printerResp.Temperature.Tool0.Actual,
		HotendTarget: printerResp.Temperature.Tool0.Target,
	}

	// If printing, fetch job info
	if status.Status == "printing" {
		jobResp, err := client.GetJob()
		if err == nil && jobResp != nil {
			status.Progress = &models.ProgressInfo{
				Completion:     jobResp.Progress.Completion,
				PrintTime:      jobResp.Progress.PrintTime,
				PrintTimeLeft:  jobResp.Progress.PrintTimeLeft,
				EstimatedTotal: int(jobResp.Job.EstimatedPrintTime),
				FileName:       jobResp.Job.File.Display,
				FilamentLength: jobResp.Job.Filament.Tool0.Length,
			}

			// Get thumbnail URL
			if jobResp.Job.File.Path != "" {
				status.ThumbnailURL = client.GetThumbnail(jobResp.Job.File.Path)
			}
		}
	}

	// Fetch current spool
	spoolID, err := client.GetCurrentSpool(0)
	if err == nil && spoolID != "" {
		spool, err := h.spoolmanClient.GetSpool(spoolID)
		if err == nil && spool != nil {
			status.CurrentSpool = spoolman.FormatSpoolInfo(spool)
		}
	}

	return status
}
