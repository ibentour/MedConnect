#!/bin/bash

# Define paths
ROOT_DIR=$(pwd)
FRONTEND_DIR="$ROOT_DIR/frontend"
BACKEND_DIR="$ROOT_DIR/backend"

echo "======================================================"
echo " 🚀 Starting up MedConnect Oriental Local Environment "
echo "======================================================"
echo ""

# 1. Set up trapped cleanup function upon Ctrl+C
cleanup() {
    echo ""
    echo "🛑 Shutting down frontend and backend processes..."
    kill $BACKEND_PID $FRONTEND_PID 2>/dev/null
    
    echo "🐳 Do you want to stop the Docker containers too? (y/N)"
    read -t 5 -n 1 answer
    if [[ $answer =~ ^[Yy]$ ]]; then
        echo ""
        docker-compose down
    fi
    
    echo ""
    echo "👋 MedConnect local environment stopped safely."
    exit 0
}

trap cleanup SIGINT SIGTERM

# 2. Bring up Database and External Services via Docker
echo "🐳 1/3 Starting Docker containers (Postgres, Ollama AI, Evolution API)..."
docker-compose up -d

# 3. Start Backend server
echo "⚙️  2/3 Starting Go API Backend (Port 3000)..."
cd "$BACKEND_DIR" || exit
go run ./cmd/server/main.go &
BACKEND_PID=$!

# Wait exactly 2 seconds to let the backend fully bind to the port
sleep 2

# 4. Start Frontend
echo "💻 3/3 Starting React Frontend (Port 5173)..."
cd "$FRONTEND_DIR" || exit
npm run dev &
FRONTEND_PID=$!

echo ""
echo "✅ Architecture is online!"
echo "➡️  Frontend available at: http://localhost:5173"
echo "➡️  API Backend available at: http://localhost:3000"
echo "Press Ctrl+C at any time to softly shutdown the entire stack."

# Keep main script alive to listen for trap
wait
