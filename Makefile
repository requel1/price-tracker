include .env
export

DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
.PHONY: help migrate-up migrate-down migrate-create migrate-force migrate-version

help:
	@echo "Доступные команды:"
	@echo "  make migrate-up              — накатить все миграции"
	@echo "  make migrate-down            — откатить одну миграцию"
	@echo "  make migrate-create name=X   — создать миграцию X"
	@echo "  make migrate-force version=N — принудительно установить версию N"
	@echo "  make migrate-version         — текущая версия миграций"

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-force:
	migrate -path migrations -database "$(DB_URL)" force $(version)

migrate-version:
	migrate -path migrations -database "$(DB_URL)" version