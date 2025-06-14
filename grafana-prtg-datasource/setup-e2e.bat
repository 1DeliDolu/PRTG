@echo off
echo Setting up Grafana for e2e testing...

REM Create necessary directories
if not exist "playwright\.auth" mkdir playwright\.auth

REM Check if Grafana is running
curl -s http://localhost:3001/api/health >nul 2>&1
if errorlevel 1 (
    echo Grafana is not running on localhost:3001
    echo Please start Grafana first:
    echo   - Using Docker: docker run -d -p 3001:3000 grafana/grafana:latest
    echo   - Or install locally and run: grafana-server
    exit /b 1
)

echo Grafana is running on localhost:3001

REM Try to login and check if admin/admin works
echo Testing admin credentials...
curl -s -o nul -w "%%{http_code}" -X POST http://localhost:3001/login -H "Content-Type: application/json" -d "{\"user\":\"admin\",\"password\":\"admin\"}" > temp_response.txt
set /p response=<temp_response.txt
del temp_response.txt

if "%response%"=="200" (
    echo ✅ Admin credentials admin/admin are working
) else if "%response%"=="302" (
    echo ✅ Admin credentials admin/admin are working
) else (
    echo ❌ Admin credentials failed HTTP %response%
    echo.
    echo To fix this:
    echo 1. Reset Grafana admin password:
    echo    grafana-cli admin reset-admin-password admin
    echo.
    echo 2. If using Docker, try:
    echo    docker run -d -p 3001:3000 -e "GF_SECURITY_ADMIN_PASSWORD=admin" grafana/grafana:latest
)

REM Install dependencies if not already installed
if not exist "node_modules" (
    echo Installing dependencies...
    npm install
)

REM Install playwright browsers
echo Installing Playwright browsers...
npx playwright install chromium

echo.
echo Setup complete! You can now run 'npm run e2e' to execute tests.
echo Make sure Grafana admin credentials are: admin/admin