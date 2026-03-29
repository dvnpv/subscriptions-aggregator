APP_NAME=subscription-aggregator

run:
	go run ./cmd/app

build:
	go build -o $(APP_NAME) ./cmd/app

test:
	go test ./...

swag:
	swag init -g cmd/app/main.go -o docs

migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/subscriptions?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/subscriptions?sslmode=disable" down

docker-up:
	docker compose up --build
    
smoke:
	chmod +x ./scripts/smoke.sh && ./scripts/smoke.sh

docker-down:
	docker compose down -v
