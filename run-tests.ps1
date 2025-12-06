# Wrapper script for running Envoy tests
# Usage: .\run-tests.ps1
# This script fixes the issue where ./test-envoy.ps1 opens "Mouse Properties" in Windows 11

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  GoBigTech Envoy Test Runner" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Get the script directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

# Check if test-envoy.ps1 exists
$testScript = Join-Path $scriptDir "test-envoy.ps1"
if (-not (Test-Path $testScript)) {
    Write-Host "Error: test-envoy.ps1 not found in $scriptDir" -ForegroundColor Red
    Write-Host "Current directory: $(Get-Location)" -ForegroundColor Yellow
    exit 1
}

Write-Host "Found test script: $testScript" -ForegroundColor Green
Write-Host ""

# Run the test script using explicit call operator
try {
    & $testScript @args
} catch {
    Write-Host "Error running test script: $_" -ForegroundColor Red
    Write-Host ""
    Write-Host "Trying alternative method..." -ForegroundColor Yellow
    powershell -ExecutionPolicy Bypass -File $testScript @args
}

