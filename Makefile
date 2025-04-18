include scripts/*.mk

DEV_COMPOSE_ARGS=--env-file config/.env.dev -f docker-compose.dev.yaml
DEV_COMPOSE_ENV=docker compose $(DEV_COMPOSE_ARGS)
DEV_COMPOSE=docker compose $(DEV_COMPOSE_ARGS)

dev-build: dev-reader-build dev-api-build
	$(DEV_COMPOSE) build

dev-up: dev-build dev-reader-up dev-api-up
	$(DEV_COMPOSE) --env-file config/.env.dev up -d

dev-migrate-up:
	docker-compose -f docker-compose.dev.yaml up -d migrations-up

dev-migrate-down:
	docker compose --profile migrations-down -f docker-compose.dev.yaml up -d migrations-down

dev-reader-build: reader_docker_build

dev-api-build: api_docker_build

dev-reader-up: 
	$(DEV_COMPOSE) -f docker-compose.dev.yaml up -d reader-up

dev-api-up:
	$(DEV_COMPOSE) -f docker-compose.dev.yaml up -d api-up

redis-connect:
	docker exec -it average_incomes.redis redis-cli -p 6379