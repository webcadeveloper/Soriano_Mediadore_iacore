@echo off
echo ========================================
echo   INICIANDO FRONTEND ANGULAR
echo   Puerto: 4200
echo ========================================
echo.

cd /d "%~dp0"

echo Iniciando servidor Angular en http://localhost:4200...
echo.

call npm start

pause
