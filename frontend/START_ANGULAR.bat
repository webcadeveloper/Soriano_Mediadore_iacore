@echo off
title Soriano Mediadores - Angular Dev Server
echo ========================================
echo   INICIANDO SERVIDOR ANGULAR
echo ========================================
echo.
echo Puerto: 4200
echo URL: http://localhost:4200
echo.
echo Compilando... (puede tardar 1-2 minutos)
echo.

cd /d "%~dp0"
"C:\Program Files\nodejs\npm.cmd" start

pause
