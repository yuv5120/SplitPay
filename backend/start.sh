#!/bin/bash

# Start Script for Split-It Backend (Go)
# Usage: ./start.sh

echo "🚀 Starting Split-It Backend (Golang)..."
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed!"
    echo "Please install Go from: https://go.dev/dl/"
    exit 1
fi

echo "✅ Go version: $(go version)"

# Check if .env exists
if [ ! -f .env ]; then
    echo "⚠️  .env file not found!"
    echo "Creating .env from .env.example..."
    cp .env.example .env
    echo "⚠️  Please edit .env with your configuration"
    exit 1
fi

# Check if Firebase service account exists
if [ ! -f firebase-service-account.json ]; then
    echo "⚠️  firebase-service-account.json not found!"
    echo "Please download from Firebase Console and save in this directory"
    exit 1
fi

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)

echo "📊 Configuration:"
echo "   PORT: ${PORT:-5000}"
echo "   MONGODB_URI: ${MONGODB_URI}"
echo "   CLIENT_URL: ${CLIENT_URL}"
echo ""

# Check if MongoDB is accessible (optional)
echo "🔍 Checking MongoDB connection..."
if command -v mongosh &> /dev/null; then
    if mongosh --eval "db.adminCommand('ping')" --quiet &> /dev/null; then
        echo "✅ MongoDB is running"
    else
        echo "⚠️  MongoDB is not accessible"
        echo "   Make sure MongoDB is running: brew services start mongodb-community"
    fi
else
    echo "⚠️  mongosh not found - skipping MongoDB check"
fi

echo ""
echo "🎯 Starting server on port ${PORT:-5000}..."
echo "   Health check: http://localhost:${PORT:-5000}/health"
echo "   API: http://localhost:${PORT:-5000}/api"
echo ""
echo "Press Ctrl+C to stop"
echo ""

# Start the server
go run main.go
