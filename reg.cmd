@echo off
setlocal enabledelayedexpansion

:: Check for administrator privileges.
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo Requesting administrator privileges...
    powershell -Command "Start-Process '%~f0' -Verb RunAs"
    exit /b
)

cls
echo Running as Administrator...
echo.

:: Get the install directory.
set "INSTALL_DIR=%~dp0"
set "INSTALL_DIR=%INSTALL_DIR:~0,-1%"

echo Target install directory: %INSTALL_DIR%
echo.

:: Making changes to the registry.
echo Adding registry entries...
set "HGL_KEY_PATH=HKLM\SOFTWARE\WOW6432Node\Flagship Studios\Hellgate London"

:: Create the key if it doesn't exist.
reg add "%HGL_KEY_PATH%" /f >nul
if %errorlevel% neq 0 (
    echo [ERROR] Failed to create registry key: %HGL_KEY_PATH%
    pause
    exit /b 1
)

:: Add HellgateCUKey.
reg add "%HGL_KEY_PATH%" /v HellgateCUKey /t REG_SZ /d "%INSTALL_DIR%" /f
if %errorlevel% neq 0 (
    echo [ERROR] Failed to add HellgateCUKey
    pause
    exit /b 1
) else (
    echo [SUCCESS] HellgateCUKey added
)

:: Add HellgateKey.
reg add "%HGL_KEY_PATH%" /v HellgateKey /t REG_SZ /d "%INSTALL_DIR%" /f
if %errorlevel% neq 0 (
    echo [ERROR] Failed to add HellgateKey
    pause
    exit /b 1
) else (
    echo [SUCCESS] HellgateKey added
)

echo.

pause