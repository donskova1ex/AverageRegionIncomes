include scripts/*.mk

DEV_COMPOSE_ARGS=--env-file .env.dev -f docker-compose.dev.yaml
DEV_COMPOSE_ENV=docker compose $(DEV_COMPOSE_ARGS)
DEV_COMPOSE=docker compose $(DEV_COMPOSE_ARGS)

dev-build: dev-reader-build
	$(DEV_COMPOSE) build

dev-up: dev-build dev-reader-up
	$(DEV_COMPOSE) --env-file .env.dev up -d

dev-migrate-up:
	docker-compose -f docker-compose.dev.yaml up -d migrations-up

dev-migrate-down:
	docker compose --profile migrations-down -f docker-compose.dev.yaml up -d migrations-down

dev-reader-build: reader_docker_build

dev-reader-up: 
	$(DEV_COMPOSE) -f docker-compose.dev.yaml up -d reader-up