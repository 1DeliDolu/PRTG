#!/bin/bash

echo "Starting Grafana with Docker for e2e testing..."

# Stop any existing Grafana container
docker-compose down

# Start Grafana with proper configuration
docker-compose up -d

# Wait for Grafana to be ready
echo "Waiting for Grafana to start..."
sleep 10

# Check if Grafana is ready
max_attempts=30
attempt=1

while [ $attempt -le $max_attempts ]; do
    if curl -s http://localhost:3001/api/health > /dev/null; then
        echo "✅ Grafana is ready at http://localhost:3001"
        echo "Admin credentials: admin/admin"
        break
    else
        echo "Waiting for Grafana... (attempt $attempt/$max_attempts)"
        sleep 2
        attempt=$((attempt + 1))
    fi
done

if [ $attempt -gt $max_attempts ]; then
    echo "❌ Grafana failed to start after $max_attempts attempts"
    exit 1
fi

echo ""
echo "You can now run the e2e tests with: npm run e2e"
echo "Or set up the testing environment with: ./setup-e2e.sh"