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

document.addEventListener('alpine:init', () => {
    Alpine.data('dashboard', () => ({
        loading: true,
        error: null,
        printers: [],
        showReturnOverlay: false,
        updateInterval: null,

        async init() {
            console.log('Initializing OctoDash...');
            
            // Set up printers from config
            this.printers = PRINTERS || [];
            
            // Start fetching status
            await this.fetchStatus();
            
            // Set up polling every second
            this.updateInterval = setInterval(() => {
                this.fetchStatus();
            }, 1000);
            
            this.loading = false;
        },

        async fetchStatus() {
            try {
                const response = await fetch('/api/status');
                if (!response.ok) {
                    throw new Error('Failed to fetch status');
                }
                
                const data = await response.json();
                
                // Update printer data
                if (data.printers) {
                    // Merge new data with existing to maintain order
                    this.printers = this.printers.map(printer => {
                        const updated = data.printers.find(p => p.id === printer.id);
                        return updated || printer;
                    });
                }
            } catch (err) {
                console.error('Error fetching status:', err);
                // Don't show error on every failed poll
                if (this.loading) {
                    this.error = 'Failed to connect to printers';
                }
            }
        },

        openPrinter(printer) {
            console.log('Opening printer:', printer.name);
            
            // Clear the update interval
            if (this.updateInterval) {
                clearInterval(this.updateInterval);
            }
            
            // Navigate to OctoPrint URL
            window.location.href = printer.octoprint_url;
        },

        returnToDashboard() {
            // This will be called from OctoPrint via parent frame
            // For now, just reload the dashboard
            window.location.href = '/';
        },

        // Formatting functions
        formatStatus(status) {
            const statusMap = {
                'idle': 'Ready',
                'printing': 'Printing',
                'error': 'Error',
                'offline': 'Offline'
            };
            return statusMap[status] || status || 'Unknown';
        },

        formatTime(seconds) {
            if (!seconds || seconds <= 0) {
                return '--:--:--';
            }

            const hours = Math.floor(seconds / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            const secs = seconds % 60;

            if (hours > 0) {
                return `${hours}h ${minutes}m ${secs}s`;
            } else if (minutes > 0) {
                return `${minutes}m ${secs}s`;
            }
            return `${secs}s`;
        },

        formatTemp(actual, target) {
            if (actual === undefined || actual === null) {
                return '--°C';
            }
            
            const actualRounded = Math.round(actual);
            const targetRounded = Math.round(target || 0);
            
            if (targetRounded > 0) {
                return `${actualRounded}°C / ${targetRounded}°C`;
            }
            return `${actualRounded}°C`;
        },

        formatWeight(grams) {
            if (!grams || grams <= 0) {
                return '--';
            }
            return `${Math.round(grams)}g`;
        },

        // Clean up on page unload
        destroy() {
            if (this.updateInterval) {
                clearInterval(this.updateInterval);
            }
        }
    }));
});

// Handle cleanup
window.addEventListener('beforeunload', () => {
    const dashboardEl = document.querySelector('[x-data="dashboard"]');
    if (dashboardEl && dashboardEl.__x) {
        const dashboard = dashboardEl.__x.$data;
        if (dashboard && dashboard.destroy) {
            dashboard.destroy();
        }
    }
});

// Listen for escape key to return to dashboard
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        // If we're in an iframe or different page, navigate back
        window.location.href = '/';
    }
});

// Add visibility change handler to pause/resume updates
document.addEventListener('visibilitychange', () => {
    const dashboardEl = document.querySelector('[x-data="dashboard"]');
    if (dashboardEl && dashboardEl.__x) {
        const dashboard = dashboardEl.__x.$data;
        
        if (document.hidden) {
            // Page is hidden, stop updates
            if (dashboard.updateInterval) {
                clearInterval(dashboard.updateInterval);
                dashboard.updateInterval = null;
            }
        } else {
            // Page is visible again, resume updates
            if (!dashboard.updateInterval) {
                dashboard.fetchStatus();
                dashboard.updateInterval = setInterval(() => {
                    dashboard.fetchStatus();
                }, 1000);
            }
        }
    }
});