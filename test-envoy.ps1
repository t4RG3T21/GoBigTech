# Test script for GoBigTech Envoy API Gateway
# Tests all services through Envoy and validates authentication, routing, and direct access restrictions
#
# Usage:
#   .\test-envoy.ps1
#   .\test-envoy.ps1 -Verbose
#   powershell -ExecutionPolicy Bypass -File .\test-envoy.ps1

param(
    [string]$EnvoyHost = "localhost",
    [int]$EnvoyPort = 80,
    [int]$EnvoyGrpcPort = 9091,
    [int]$EnvoyGrpcRestPort = 8084,
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"
$global:TestResults = @()
$global:PassedTests = 0
$global:FailedTests = 0

# Colors for output
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Error { Write-Host $args -ForegroundColor Red }
function Write-Info { Write-Host $args -ForegroundColor Cyan }
function Write-Warning { Write-Host $args -ForegroundColor Yellow }

# Test result tracking
function Add-TestResult {
    param(
        [string]$TestName,
        [bool]$Passed,
        [string]$Message = ""
    )
    
    $result = @{
        Name = $TestName
        Passed = $Passed
        Message = $Message
        Timestamp = Get-Date
    }
    
    $global:TestResults += $result
    
    if ($Passed) {
        $global:PassedTests++
        Write-Success "✓ $TestName"
        if ($Message -and $Verbose) {
            Write-Info "  $Message"
        }
    } else {
        $global:FailedTests++
        Write-Error "✗ $TestName"
        if ($Message) {
            Write-Error "  $Message"
        }
    }
}

# HTTP request helper
function Invoke-EnvoyRequest {
    param(
        [string]$Method = "GET",
        [string]$Path,
        [hashtable]$Headers = @{},
        [object]$Body = $null,
        [int]$ExpectedStatus = 200
    )
    
    $url = "http://${EnvoyHost}:${EnvoyPort}${Path}"
    
    try {
        $params = @{
            Method = $Method
            Uri = $url
            Headers = $Headers
            ContentType = "application/json"
            TimeoutSec = 10
        }
        
        if ($Body) {
            $params.Body = ($Body | ConvertTo-Json -Depth 10)
        }
        
        $response = Invoke-WebRequest @params -UseBasicParsing -ErrorAction Stop
        
        return @{
            Success = ($response.StatusCode -eq $ExpectedStatus)
            StatusCode = $response.StatusCode
            Content = $response.Content | ConvertFrom-Json -ErrorAction SilentlyContinue
            RawContent = $response.Content
            Headers = $response.Headers
        }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        return @{
            Success = ($statusCode -eq $ExpectedStatus)
            StatusCode = $statusCode
            Content = $null
            RawContent = $_.Exception.Response | Out-String
            Error = $_.Exception.Message
        }
    }
}

# Check service health
function Test-ServiceHealth {
    Write-Info "`n=== Testing Service Health ==="
    
    # Test Envoy health
    try {
        $response = Invoke-WebRequest -Uri "http://${EnvoyHost}:9901/server_info" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
        Add-TestResult -TestName "Envoy Admin Interface" -Passed $true -Message "Envoy admin interface is accessible"
    } catch {
        Add-TestResult -TestName "Envoy Admin Interface" -Passed $false -Message "Cannot access Envoy admin: $($_.Exception.Message)"
    }
    
    # Test health endpoint through Envoy
    $healthResponse = Invoke-EnvoyRequest -Path "/health" -ExpectedStatus 200
    Add-TestResult -TestName "Health Check via Envoy" -Passed $healthResponse.Success -Message "Status: $($healthResponse.StatusCode)"
    
    # Test IAM service through Envoy (public endpoint)
    # Envoy routes /api/iam/* to iam-service:8082 which handles /v1/iam/*
    # So full path through Envoy is /api/iam/v1/iam/register
    $iamResponse = Invoke-EnvoyRequest -Path "/api/iam/v1/iam/register" -Method "POST" -Body @{
        email = "test@example.com"
        password = "testpass123"
        username = "testuser"
    } -ExpectedStatus 200
    Add-TestResult -TestName "IAM Service via Envoy" -Passed ($iamResponse.Success -or $iamResponse.StatusCode -eq 409) -Message "Status: $($iamResponse.StatusCode)"
}

# Test authentication flow
function Test-AuthenticationFlow {
    Write-Info "`n=== Testing Authentication Flow ==="
    
    $testEmail = "test-$(Get-Date -Format 'yyyyMMddHHmmss')@example.com"
    $testPassword = "TestPass123!"
    $testUsername = "testuser-$(Get-Date -Format 'yyyyMMddHHmmss')"
    
    # Step 1: Register user
    Write-Info "Step 1: Registering new user..."
    $registerBody = @{
        email = $testEmail
        password = $testPassword
        username = $testUsername
    }
    
    $registerResponse = Invoke-EnvoyRequest -Path "/api/iam/v1/iam/register" -Method "POST" -Body $registerBody -ExpectedStatus 200
    
    if (-not $registerResponse.Success) {
        # User might already exist, try login instead
        if ($registerResponse.StatusCode -eq 409) {
            Write-Warning "User already exists, proceeding with login..."
        } else {
            Add-TestResult -TestName "User Registration" -Passed $false -Message "Status: $($registerResponse.StatusCode), Error: $($registerResponse.Error)"
            return $null
        }
    } else {
        Add-TestResult -TestName "User Registration" -Passed $true -Message "User registered successfully"
    }
    
    # Step 2: Login
    Write-Info "Step 2: Logging in..."
    $loginBody = @{
        email = $testEmail
        password = $testPassword
    }
    
    $loginResponse = Invoke-EnvoyRequest -Path "/api/iam/v1/iam/login" -Method "POST" -Body $loginBody -ExpectedStatus 200
    
    if (-not $loginResponse.Success) {
        Add-TestResult -TestName "User Login" -Passed $false -Message "Status: $($loginResponse.StatusCode), Error: $($loginResponse.Error)"
        return $null
    }
    
    Add-TestResult -TestName "User Login" -Passed $true -Message "Login successful"
    
    # Extract token
    $token = $null
    if ($loginResponse.Content -and $loginResponse.Content.session_id) {
        $token = $loginResponse.Content.session_id
    } elseif ($loginResponse.Content -and $loginResponse.Content.token) {
        $token = $loginResponse.Content.token
    } else {
        Add-TestResult -TestName "Token Extraction" -Passed $false -Message "No token in response: $($loginResponse.RawContent)"
        return $null
    }
    
    Add-TestResult -TestName "Token Extraction" -Passed ($null -ne $token) -Message "Token extracted: $($token.Substring(0, [Math]::Min(20, $token.Length)))..."
    
    return $token
}

# Test order creation with authentication
function Test-OrderCreation {
    param([string]$Token)
    
    Write-Info "`n=== Testing Order Creation with Authentication ==="
    
    if (-not $Token) {
        Add-TestResult -TestName "Order Creation (Authenticated)" -Passed $false -Message "No token provided"
        return
    }
    
    # Create order with token
    $orderBody = @{
        user_id = "test-user-123"
        items = @(
            @{
                product_id = "prod-123"
                quantity = 2
            },
            @{
                product_id = "prod-456"
                quantity = 1
            }
        )
    }
    
    $headers = @{
        "Authorization" = "Bearer $Token"
    }
    
    $orderResponse = Invoke-EnvoyRequest -Path "/api/orders" -Method "POST" -Body $orderBody -Headers $headers -ExpectedStatus 200
    
    if ($orderResponse.Success) {
        Add-TestResult -TestName "Order Creation (Authenticated)" -Passed $true -Message "Order created successfully"
        if ($orderResponse.Content -and $orderResponse.Content.id) {
            Write-Info "  Order ID: $($orderResponse.Content.id)"
            return $orderResponse.Content.id
        }
    } else {
        # 201 is also acceptable for creation
        if ($orderResponse.StatusCode -eq 201) {
            Add-TestResult -TestName "Order Creation (Authenticated)" -Passed $true -Message 'Order created (201 Created)'
        } else {
            Add-TestResult -TestName "Order Creation (Authenticated)" -Passed $false -Message "Status: $($orderResponse.StatusCode), Error: $($orderResponse.Error)"
        }
    }
    
    return $null
}

# Test unauthorized access
function Test-UnauthorizedAccess {
    Write-Info "`n=== Testing Unauthorized Access (Should Return 401) ==="
    
    # Try to create order without token
    $orderBody = @{
        user_id = "test-user-123"
        items = @(
            @{
                product_id = "prod-123"
                quantity = 2
            }
        )
    }
    
    $orderResponse = Invoke-EnvoyRequest -Path "/api/orders" -Method "POST" -Body $orderBody -ExpectedStatus 401
    
    Add-TestResult -TestName "Unauthorized Access (No Token)" -Passed $orderResponse.Success -Message "Status: $($orderResponse.StatusCode) (expected 401)"
    
    # Try with invalid token
    $invalidHeaders = @{
        "Authorization" = "Bearer invalid-token-12345"
    }
    
    $invalidResponse = Invoke-EnvoyRequest -Path "/api/orders" -Method "POST" -Body $orderBody -Headers $invalidHeaders -ExpectedStatus 401
    
    Add-TestResult -TestName "Unauthorized Access (Invalid Token)" -Passed $invalidResponse.Success -Message "Status: $($invalidResponse.StatusCode) (expected 401)"
}

# Test gRPC routing through Envoy
function Test-GrpcRouting {
    Write-Info "`n=== Testing gRPC Routing through Envoy ==="
    
    # Check if grpcurl is available
    $grpcurlAvailable = $false
    try {
        $null = Get-Command grpcurl -ErrorAction Stop
        $grpcurlAvailable = $true
    } catch {
        Write-Warning "grpcurl not found. Skipping native gRPC tests."
    }
    
    if ($grpcurlAvailable) {
        # Test gRPC service discovery
        try {
            $grpcResponse = grpcurl -plaintext "${EnvoyHost}:${EnvoyGrpcPort}" list 2>&1
            if ($LASTEXITCODE -eq 0) {
                Add-TestResult -TestName "gRPC Service Discovery" -Passed $true -Message "gRPC services discovered"
                if ($Verbose) {
                    Write-Info "  Available services: $grpcResponse"
                }
            } else {
                Add-TestResult -TestName "gRPC Service Discovery" -Passed $false -Message "Failed to discover gRPC services"
            }
        } catch {
            Add-TestResult -TestName "gRPC Service Discovery" -Passed $false -Message "Error: $($_.Exception.Message)"
        }
    }
    
    # Test REST→gRPC transformation (Inventory Service)
    Write-Info "Testing REST→gRPC transformation for Inventory Service..."
    $inventoryBody = @{
        product_id = "prod-123"
    }
    
    $inventoryResponse = Invoke-EnvoyRequest -Path "/inventory.v1.InventoryService/GetStock" -Method "POST" -Body $inventoryBody -ExpectedStatus 200
    
    if ($inventoryResponse.Success -or $inventoryResponse.StatusCode -eq 200) {
        Add-TestResult -TestName "REST→gRPC (Inventory Service)" -Passed $true -Message "Status: $($inventoryResponse.StatusCode)"
    } else {
        # 404 or 503 might be acceptable if service is not fully ready
        if ($inventoryResponse.StatusCode -in @(404, 503, 502)) {
            Add-TestResult -TestName "REST→gRPC (Inventory Service)" -Passed $false -Message "Service might not be ready (Status: $($inventoryResponse.StatusCode))"
        } else {
            Add-TestResult -TestName "REST→gRPC (Inventory Service)" -Passed $false -Message "Status: $($inventoryResponse.StatusCode)"
        }
    }
}

# Test direct service access (should fail)
function Test-DirectAccessRestriction {
    Write-Info "`n=== Testing Direct Service Access Restrictions ==="
    
    $services = @(
        @{ Name = "Order Service"; Port = 8080 }
        @{ Name = "IAM Service"; Port = 8082 }
        @{ Name = "Notification Service"; Port = 8083 }
    )
    
    foreach ($service in $services) {
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:$($service.Port)/health" -UseBasicParsing -TimeoutSec 2 -ErrorAction Stop
            # If we get here, service is accessible directly (might be OK in dev, but should be restricted in prod)
            Add-TestResult -TestName "Direct Access: $($service.Name)" -Passed $false -Message "Service is accessible directly on port $($service.Port) (should be restricted)"
        } catch {
            # Connection refused or timeout is expected
            if ($_.Exception.Message -match "refused|timeout|unreachable") {
                Add-TestResult -TestName "Direct Access: $($service.Name)" -Passed $true -Message "Service correctly restricted (not accessible directly)"
            } else {
                Add-TestResult -TestName "Direct Access: $($service.Name)" -Passed $false -Message "Unexpected error: $($_.Exception.Message)"
            }
        }
    }
}

# Test public endpoints (should work without auth)
function Test-PublicEndpoints {
    Write-Info "`n=== Testing Public Endpoints (No Auth Required) ==="
    
    # Health endpoint
    $healthResponse = Invoke-EnvoyRequest -Path "/health" -ExpectedStatus 200
    Add-TestResult -TestName "Public Endpoint: /health" -Passed $healthResponse.Success -Message "Status: $($healthResponse.StatusCode)"
    
    # IAM register endpoint (public)
    $registerBody = @{
        email = "public-test-$(Get-Date -Format 'yyyyMMddHHmmss')@example.com"
        password = "TestPass123!"
        username = "publictest"
    }
    
    # Envoy routes /api/iam/* to iam-service:8082 which handles /v1/iam/*
    $registerResponse = Invoke-EnvoyRequest -Path "/api/iam/v1/iam/register" -Method "POST" -Body $registerBody -ExpectedStatus 200
    Add-TestResult -TestName "Public Endpoint: /api/iam/v1/iam/register" -Passed ($registerResponse.Success -or $registerResponse.StatusCode -eq 409) -Message "Status: $($registerResponse.StatusCode)"
    
    # IAM login endpoint (public)
    $loginBody = @{
        email = "public-test@example.com"
        password = "TestPass123!"
    }
    
    $loginResponse = Invoke-EnvoyRequest -Path "/api/iam/v1/iam/login" -Method "POST" -Body $loginBody -ExpectedStatus 200
    # 401 is acceptable if user doesn't exist
    Add-TestResult -TestName "Public Endpoint: /api/iam/v1/iam/login" -Passed ($loginResponse.StatusCode -in @(200, 401)) -Message "Status: $($loginResponse.StatusCode)"
}

# Main execution
function Main {
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host "  GoBigTech Envoy API Gateway Tests" -ForegroundColor Cyan
    Write-Host "========================================`n" -ForegroundColor Cyan
    
    Write-Info "Configuration:"
    Write-Info "  Envoy Host: $EnvoyHost"
    Write-Info "  Envoy HTTP Port: $EnvoyPort"
    Write-Info "  Envoy gRPC Port: $EnvoyGrpcPort"
    Write-Info "  Envoy gRPC REST Port: $EnvoyGrpcRestPort"
    Write-Host ""
    
    # Run all tests
    Test-ServiceHealth
    $token = Test-AuthenticationFlow
    Test-OrderCreation -Token $token
    Test-UnauthorizedAccess
    Test-PublicEndpoints
    Test-GrpcRouting
    Test-DirectAccessRestriction
    
    # Print summary
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host "  Test Summary" -ForegroundColor Cyan
    Write-Host "========================================`n" -ForegroundColor Cyan
    
    Write-Info "Total Tests: $($global:TestResults.Count)"
    Write-Success "Passed: $global:PassedTests"
    Write-Error "Failed: $global:FailedTests"
    
    $successRate = if ($global:TestResults.Count -gt 0) {
        [math]::Round(($global:PassedTests / $global:TestResults.Count) * 100, 2)
    } else {
        0
    }
    
    Write-Info "Success Rate: $successRate%"
    
    if ($global:FailedTests -gt 0) {
        Write-Host "`nFailed Tests:" -ForegroundColor Red
        foreach ($result in $global:TestResults | Where-Object { -not $_.Passed }) {
            Write-Error "  - $($result.Name): $($result.Message)"
        }
    }
    
    Write-Host ""
    
    # Exit with appropriate code
    if ($global:FailedTests -eq 0) {
        Write-Success "All tests passed! ✓"
        exit 0
    } else {
        Write-Error "Some tests failed. Please review the output above."
        exit 1
    }
}

# Run main function
Main

