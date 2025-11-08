.PHONY: test test-coverage coverage-badge

test:
	go test -v ./... -coverprofile=coverage.out

test-coverage: test
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

coverage-badge:
	go test -coverprofile=coverage.out ./...
	coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Покрытие тестами: $${coverage}%"