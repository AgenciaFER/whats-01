#!/bin/bash

# Check if .env exists, if not create it from example
if [ ! -f .env ]; then
    echo "Creating .env file from example..."
    cp env.example .env
fi

# Ensure store directory exists
STORE_DIR="${STORE_DIR:-$HOME/.whatsapp-panel}"
mkdir -p "$STORE_DIR"

# Download dependencies if needed
go mod download

# Run the server
go run cmd/server/main.go