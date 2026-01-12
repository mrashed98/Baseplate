.PHONY: build run test clean db-up db-down db-reset migrate

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run the application
run:
	go run ./cmd/server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Start the database
db-up:
	docker-compose up -d db

# Stop the database
db-down:
	docker-compose down

# Reset database (drop and recreate)
db-reset:
	docker-compose down -v
	docker-compose up -d db
	sleep 3
	@echo "Database reset complete"

# Run migrations manually
migrate:
	docker exec -i baseplate_db psql -U user -d baseplate < migrations/001_initial.sql

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate swagger docs (requires swag)
swagger:
	swag init -g cmd/server/main.go -o docs/swagger

# Tidy dependencies
tidy:
	go mod tidy
