# Copyright 2025 William Marchesi

# Author: William Marchesi
# Email: will@marchesi.io
# Website: https://marchesi.io/

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build for ARM64 (same as spool-scanner)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
    -ldflags="-w -s" \
    -o octodash \
    cmd/server/main.go

# Final stage
FROM debian:bullseye-slim

# Install ca-certificates for HTTPS
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -m -u 1001 appuser

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/octodash .

# Copy web assets
COPY --from=builder /app/web ./web

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Use non-root user
USER appuser

EXPOSE 8080

CMD ["./octodash"]