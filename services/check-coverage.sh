#!/bin/bash

# Script to check test coverage for all services
# Usage: ./check-coverage.sh

set -e

echo "=== Checking Test Coverage for All Services ==="
echo ""

TOTAL_COVERAGE=0
SERVICES_COUNT=0

check_service() {
    local service=$1
    local test_path=$2
    
    echo "--- $service Service ---"
    cd "services/$service"
    
    if [ -f "go.mod" ]; then
        go test -coverprofile=coverage.out $test_path
        if [ -f "coverage.out" ]; then
            COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
            echo "Coverage: $COVERAGE"
            
            # Extract percentage number
            COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')
            
            if (( $(echo "$COVERAGE_NUM < 40" | bc -l) )); then
                echo "❌ Coverage is below 40%: $COVERAGE"
                cd ../..
                return 1
            else
                echo "✅ Coverage is acceptable: $COVERAGE"
                TOTAL_COVERAGE=$(echo "$TOTAL_COVERAGE + $COVERAGE_NUM" | bc)
                SERVICES_COUNT=$((SERVICES_COUNT + 1))
            fi
        else
            echo "❌ No coverage file generated"
            cd ../..
            return 1
        fi
    else
        echo "❌ No go.mod found"
        cd ../..
        return 1
    fi
    
    cd ../..
    echo ""
}

# Check Order service
check_service "order" "./..."

# Check Inventory service
check_service "inventory" "./cmd/inventory"

# Check Payment service
check_service "payment" "./cmd/payment"

# Calculate average
if [ $SERVICES_COUNT -gt 0 ]; then
    AVG_COVERAGE=$(echo "scale=2; $TOTAL_COVERAGE / $SERVICES_COUNT" | bc)
    echo "=== Summary ==="
    echo "Average Coverage: ${AVG_COVERAGE}%"
    echo "All services have at least 40% coverage: ✅"
else
    echo "❌ No services passed coverage check"
    exit 1
fi

