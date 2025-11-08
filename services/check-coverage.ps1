# PowerShell script to check test coverage for all services
# Usage: .\check-coverage.ps1

$ErrorActionPreference = "Stop"

Write-Host "=== Checking Test Coverage for All Services ===" -ForegroundColor Cyan
Write-Host ""

$totalCoverage = 0
$servicesCount = 0
$failedServices = @()

function Check-Service {
    param(
        [string]$ServiceName,
        [string]$TestPath
    )
    
    Write-Host "--- $ServiceName Service ---" -ForegroundColor Yellow
    Push-Location "services\$ServiceName"
    
    try {
        if (Test-Path "go.mod") {
            # Run tests and generate coverage
            if ($TestPath -eq "./...") {
                go test -coverprofile=coverage.out ./...
            } else {
                go test -coverprofile=coverage.out $TestPath
            }
            
            if (Test-Path "coverage.out") {
                # Get coverage percentage
                $coverageOutput = go tool cover -func=coverage.out | Select-String "total"
                $coverageLine = $coverageOutput.Line
                $coverageMatch = [regex]::Match($coverageLine, '(\d+\.\d+)%')
                
                if ($coverageMatch.Success) {
                    $coverage = [double]$coverageMatch.Groups[1].Value
                    Write-Host "Coverage: $coverage%"
                    
                    if ($coverage -lt 40) {
                        Write-Host "❌ Coverage is below 40%: $coverage%" -ForegroundColor Red
                        $script:failedServices += $ServiceName
                        Pop-Location
                        return $false
                    } else {
                        Write-Host "✅ Coverage is acceptable: $coverage%" -ForegroundColor Green
                        $script:totalCoverage += $coverage
                        $script:servicesCount++
                    }
                } else {
                    Write-Host "❌ Could not parse coverage" -ForegroundColor Red
                    $script:failedServices += $ServiceName
                    Pop-Location
                    return $false
                }
            } else {
                Write-Host "❌ No coverage file generated" -ForegroundColor Red
                $script:failedServices += $ServiceName
                Pop-Location
                return $false
            }
        } else {
            Write-Host "❌ No go.mod found" -ForegroundColor Red
            $script:failedServices += $ServiceName
            Pop-Location
            return $false
        }
    } catch {
        Write-Host "❌ Error: $_" -ForegroundColor Red
        $script:failedServices += $ServiceName
        Pop-Location
        return $false
    }
    
    Pop-Location
    Write-Host ""
    return $true
}

# Check Order service
Check-Service "order" "./..."

# Check Inventory service
Check-Service "inventory" "./cmd/inventory"

# Check Payment service
Check-Service "payment" "./cmd/payment"

# Summary
Write-Host "=== Summary ===" -ForegroundColor Cyan
if ($servicesCount -gt 0) {
    $avgCoverage = $totalCoverage / $servicesCount
    Write-Host "Average Coverage: $([math]::Round($avgCoverage, 2))%"
}

if ($failedServices.Count -eq 0) {
    Write-Host "All services have at least 40% coverage: ✅" -ForegroundColor Green
    exit 0
} else {
    Write-Host "Failed services: $($failedServices -join ', ')" -ForegroundColor Red
    Write-Host "❌ Some services have coverage below 40%" -ForegroundColor Red
    exit 1
}

