@echo off
echo ========================================
echo Testing The Campaign Server
echo ========================================
echo.

echo [1/3] Checking if server executable exists...
if exist server.exe (
    echo ✓ server.exe found
) else (
    echo ✗ server.exe not found!
    echo Please build it first: cd backend ^&^& go build -o ../server.exe ./cmd/server
    pause
    exit /b 1
)
echo.

echo [2/3] Killing any existing server instances...
taskkill /F /IM server.exe >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo ✓ Killed existing server
) else (
    echo ℹ No existing server running
)
echo.

echo [3/3] Starting server (will stop after 3 seconds)...
timeout /t 1 /nobreak >nul
start /B server.exe
timeout /t 3 /nobreak >nul
echo.

echo ========================================
echo Test Complete!
echo ========================================
echo.
echo If you saw "Server starting on :8080" above,
echo everything is working correctly!
echo.
echo To play the game:
echo 1. Run: start.bat
echo 2. Open two browser tabs to: http://localhost:8080
echo 3. Join with SAME Game ID in both tabs
echo 4. Enjoy!
echo.
echo Killing test server...
taskkill /F /IM server.exe >nul 2>&1
echo.
pause
