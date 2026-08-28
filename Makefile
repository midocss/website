DB_URL ?= postgres://$(or $(DB_USER),postgres):$(or $(DB_PASSWORD),postgres)@$(or $(DB_HOST),127.0.0.1):$(or $(DB_PORT),5432)/$(or $(DB_NAME),midocss)?sslmode=$(or $(DB_SSLMODE),disable)
MIGRATE ?= migrate -path migrations -database "$(DB_URL)"

.PHONY: run build test vet fmt tidy migrate-up migrate-down migrate-create seed infra-up infra-down

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api
	go build -o bin/seed ./cmd/seed

test:
	go test ./... -race -cover

vet:
	go vet ./...

fmt:
	gofmt -l -w .

tidy:
	go mod tidy

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

# usage: make migrate-create name=add_something
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

seed:
	go run ./cmd/seed

infra-up:
	docker compose up -d

infra-down:
	docker compose down
