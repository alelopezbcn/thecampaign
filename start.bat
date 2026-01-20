@echo off
echo ========================================
echo The Campaign - Multiplayer Card Game
echo ========================================
echo.
echo Building server...
cd backend
go build -o ..\server.exe .\cmd\server\main.go
if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)
cd ..
echo Build successful!
echo.
echo Starting server on http://localhost:8080
echo.
echo To play:
echo 1. Open TWO browser tabs
echo 2. Both navigate to: http://localhost:8080
echo 3. Enter SAME Game ID in both tabs
echo 4. Enjoy!
echo.
echo Press Ctrl+C to stop the server
echo ========================================
echo.
server.exe
