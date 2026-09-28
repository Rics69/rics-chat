include .env
export

export PROJECT_ROOT=$(shell pwd)

env-up:
	@docker compose up -d rics-chat-postgres

env-down:
	@docker compose down rics-chat-postgres

env-cleanup:
	@read -p "Очистить volume файлы бд? Опасность утери данных. [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down rics-chat-postgres port-forwarder && \
		rm -rf ${PROJECT_ROOT}/out/pgdata && \
		echo "Файлы очищены"; \
	else \
		echo "Очистка отменена"; \
	fi

env-port-forward:
	@docker compose up -d port-forwarder

env-port-close:
	@docker compose down port-forwarder

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует параметр seq. Пример: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	docker compose run --rm rics-chat-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Отсутствует параметр action. Пример: make migrate-action action=up"; \
		exit 1; \
	fi; \
	docker compose run --rm rics-chat-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@rics-chat-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"

logs-cleanup:
	@read -p "Очистить log файлы? Опасность утери логов. [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		rm -rf ${PROJECT_ROOT}/out/logs && \
		echo "Файлы логов очищены"; \
	else \
		echo "Очистка логов отменена"; \
	fi

# buf.lock фиксирует версию googleapis. Запустить один раз и закоммитить buf.lock
proto-deps:
	@docker compose run --rm buf dep update

proto-lint:
	@docker compose run --rm buf lint

# старые файлы удаляем, чтобы не остались хвосты от удалённых proto
proto-gen:
	@rm -rf ${PROJECT_ROOT}/pkg/api ${PROJECT_ROOT}/api/openapi/*.swagger.json && \
	docker compose run --rm buf generate && \
	echo "Код сгенерирован в pkg/api, swagger в api/openapi"

chat-run:
	@export LOGGER_FOLDER=${PROJECT_ROOT}/out/logs && \
	export POSTGRES_HOST=localhost && \
	go mod tidy && \
	go run ${PROJECT_ROOT}/cmd/chat

test:
	@go test -race -count=1 ./...

chat-deploy:
	@docker compose up -d --build rics-chat

chat-undeploy:
	@docker compose down rics-chat

ps:
	@docker compose ps
