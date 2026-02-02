.PHONY: zipkin graylog prometheus all stop clean run-example

# Start Zipkin
zipkin:
	docker compose up -d zipkin

# Start Graylog (includes MongoDB and OpenSearch dependencies)
graylog:
	docker compose up -d graylog

# Start Prometheus
prometheus:
	docker compose up -d prometheus

# Start all services
all:
	docker compose up -d

# Stop all services
stop:
	docker compose down

# Stop all services and remove volumes
clean:
	docker compose down -v

# Build example application
build-example:
	cd example && go build -o bin/example

# Run example Go application
run-example:
	cd example && go run .
