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

# 2. Check for port conflicts and start necessary Docker services
echo "🔍 Checking for port 5432 (Postgres)..."
if lsof -Pi :5432 -sTCP:LISTEN -t >/dev/null ; then
    echo "⚠️  [WARNING] Port 5432 is already in use by another process."
    echo "   If you have a local Postgres running, the Docker container might fail to bind."
fi

echo "🐳 1/3 Starting Docker dependencies (Postgres, Ollama AI)..."
# We start only the dependencies via Docker-compose if the intent is local dev,
# or we use --remove-orphans to ensure clean state.
docker-compose up -d postgres ollama

# Wait for database to be ready
echo "⏳ Waiting for Database to be healthy..."
for i in {1..20}; do
    if docker exec medconnect_db pg_isready -U medadmin > /dev/null 2>&1; then
        echo "✅ Database is ready!"
        break
    fi
    if [ $i -eq 20 ]; then
        echo "❌ [ERROR] Database failed to become ready in 20 seconds."
        exit 1
    fi
    sleep 1
done

# 3. Start Backend server
cd "$BACKEND_DIR" || exit
# Create logs directory if it doesn't exist
mkdir -p "$ROOT_DIR/logs"
# 3. Backend Setup & Run
echo "⚙️  2/3 Starting Go Backend API..."
# Ensure dependencies are available before starting
go mod tidy
go run ./cmd/server/main.go > "$ROOT_DIR/logs/backend.log" 2>&1 &
BACKEND_PID=$!

# Wait exactly 2 seconds to let the backend fully bind to the port
sleep 2

# 4. Frontend Setup & Run
echo "💻 3/3 Starting React Frontend..."
cd "$ROOT_DIR/frontend" || exit
# Check for node_modules
if [ ! -d "node_modules" ]; then
    echo "📦 Initializing frontend dependencies..."
    npm install
fi
npm run dev -- --host > "$ROOT_DIR/logs/frontend.log" 2>&1 &
FRONTEND_PID=$!

echo ""
echo "✅ Architecture is online!"
echo "➡️  Frontend available at: http://localhost:5173"
echo "➡️  API Backend available at: http://localhost:3000"
echo "Press Ctrl+C at any time to softly shutdown the entire stack."

# Keep main script alive to listen for trap
wait
