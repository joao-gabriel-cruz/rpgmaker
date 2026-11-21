.PHONY: up down lint test build check

up:
	docker compose up --build

down:
	docker compose down

lint:
	cd backend && golangci-lint run ./...
	cd frontend && npm run lint

test:
	cd backend && go test -p 1 ./...
	cd frontend && npm test

build:
	cd backend && go build ./...
	cd frontend && npm run build

check: lint build test
