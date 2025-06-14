#!/bin/bash

# Setup script for Grafana with proper admin credentials
echo "Setting up Grafana for e2e testing..."

# Create necessary directories
mkdir -p playwright/.auth

# Check if Grafana is running
if ! curl -s http://localhost:3001/api/health > /dev/null; then
    echo "Grafana is not running on localhost:3001"
    echo "Please start Grafana first:"
    echo "  - Using Docker: docker run -d -p 3001:3000 grafana/grafana:latest"
    echo "  - Or install locally and run: grafana-server"
    exit 1
fi

echo "Grafana is running on localhost:3001"

# Try to login and check if admin/admin works
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  http://localhost:3001/login \
  -H "Content-Type: application/json" \
  -d '{"user":"admin","password":"admin"}')

if [ "$response" = "200" ] || [ "$response" = "302" ]; then
    echo "✅ Admin credentials (admin/admin) are working"
else
    echo "❌ Admin credentials failed (HTTP $response)"
    echo ""
    echo "To fix this:"
    echo "1. Reset Grafana admin password:"
    echo "   grafana-cli admin reset-admin-password admin"
    echo ""
    echo "2. Or check your Grafana configuration:"
    echo "   - Default admin user should be 'admin'"
    echo "   - Default admin password should be 'admin'"
    echo ""
    echo "3. If using Docker, try:"
    echo "   docker run -d -p 3001:3000 -e \"GF_SECURITY_ADMIN_PASSWORD=admin\" grafana/grafana:latest"
fi

# Install dependencies if not already installed
if [ ! -d "node_modules" ]; then
    echo "Installing dependencies..."
    npm install
fi

# Install playwright browsers
echo "Installing Playwright browsers..."
npx playwright install chromium

echo ""
echo "Setup complete! You can now run 'npm run e2e' to execute tests."
echo "Make sure Grafana admin credentials are: admin/admin"