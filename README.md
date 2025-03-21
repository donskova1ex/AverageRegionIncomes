# Average Region Incomes

Сервис для обработки и анализа данных о средних доходах по регионам.

## Структура проекта

- `api/` - API endpoints и обработчики
- `cmd/` 
  - `api/` - точка входа для API сервера
  - `readers/` - сервисы для чтения данных
- `internal/`
  - `config/` - конфигурация приложения
  - `domain/` - бизнес-логика и модели
  - `middleware/` - промежуточные обработчики
  - `processors/` - обработчики данных
  - `repositories/` - слой доступа к данным
- `migrations/` - SQL миграции
- `openapi/` - OpenAPI спецификации
- `specs/` - спецификации и документация
- `files/` - файлы данных
- `db-files/` - файлы базы данных
- `config/` - конфигурационные файлы

## Технологический стек

- Go 1.23
- PostgreSQL 15
- Docker & Docker Compose
- Gorilla Mux для маршрутизации
- SQLx для работы с БД
- Excelize для работы с Excel файлами

## Компоненты системы

1. **API сервер** - REST API для доступа к данным
2. **Reader сервис** - обработка и загрузка данных
3. **База данных** - PostgreSQL для хранения обработанных данных

## Запуск проекта

### Разработка

```bash
# Запуск всех сервисов
docker-compose -f docker-compose.dev.yaml up

# Только миграции
docker-compose -f docker-compose.dev.yaml up migrations-up

# Откат миграций
docker-compose -f docker-compose.dev.yaml --profile migrations-down up migrations-down
```

### Переменные окружения

Основные переменные находятся в `config/.env.dev`:
- POSTGRES_* - настройки PostgreSQL
- API_* - настройки API сервера
- READER_* - настройки сервиса чтения данных

## Makefile

Проект включает Makefile для автоматизации основных задач разработки.

## Документация API

API документация доступна в формате OpenAPI в директории `openapi/`.
