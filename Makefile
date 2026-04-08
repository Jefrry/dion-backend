include .env

GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
GOOSE_MIGRATION_DIR=./migrations

migrate-create:
	goose -dir $(GOOSE_MIGRATION_DIR) create $(name) sql

migrate-up:
	goose -dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" up

migrate-down:
	goose -dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" down

migrate-status:
	goose -dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" status

run:
	source .env && go run ./cmd/server/main.go