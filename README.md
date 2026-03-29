# Subscriptions Aggregator

REST-сервис для управления онлайн-подписками пользователей.

## Стек

- Go 1.23
- PostgreSQL 16
- chi
- pgx
- Docker / Docker Compose
- golang-migrate
- Swagger

## Возможности

- создание подписки;
- получение подписки по ID;
- обновление подписки;
- удаление подписки;
- получение списка подписок;
- расчет суммарной стоимости подписок за период;
- Swagger UI;
- smoke test.

## Запуск через Docker Compose

```bash
docker compose up --build
```

После запуска:
- API: \`http://localhost:8080\`
- Swagger UI: \`http://localhost:8080/swagger/index.html\`

## Локальный запуск

### 1. Поднять базу

```bash
docker compose up -d db
```

### 2. Применить миграции

```bash
make migrate-up
```

### 3. Создать \`.env\`

```env
HTTP_ADDR=:8080
LOG_LEVEL=info
DATABASE_URL=postgres://postgres:postgres@localhost:5432/subscriptions?sslmode=disable
```

### 4. Запустить приложение

```bash
make run
```

## Переменные окружения

- \`HTTP_ADDR\` — адрес HTTP-сервера, например \`:8080\`
- \`LOG_LEVEL\` — уровень логирования, например \`info\`
- \`DATABASE_URL\` — строка подключения к PostgreSQL

Для локального запуска:

```text
postgres://postgres:postgres@localhost:5432/subscriptions?sslmode=disable
```

Для Docker Compose:

```text
postgres://postgres:postgres@db:5432/subscriptions?sslmode=disable
```

## API

- \`GET /health\`
- \`POST /subscriptions/\`
- \`GET /subscriptions/\`
- \`GET /subscriptions/{id}\`
- \`PUT /subscriptions/{id}\`
- \`DELETE /subscriptions/{id}\`
- \`GET /subscriptions/total\`

Swagger UI:

```text
http://localhost:8080/swagger/index.html
```

## Формат данных

Дата передается в формате:

```text
MM-YYYY
```

Пример:

```json
{
  "service_name": "Netflix",
  "price": 999,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "01-2026",
  "end_date": "03-2026"
}
```

## Примеры запросов

Создание подписки:

```bash
curl -X POST http://localhost:8080/subscriptions/ \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Netflix",
    "price": 999,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "01-2026",
    "end_date": "03-2026"
  }'
```

Получение списка:

```bash
curl http://localhost:8080/subscriptions/
```

Расчет total:

```bash
curl "http://localhost:8080/subscriptions/total?from=01-2026&to=03-2026"
```

## Smoke test

```bash
make smoke
```

## Тесты

```bash
make test
```

## Swagger

Установка \`swag\`:

```bash
brew install swaggo/swag/swag
```

Генерация документации:

```bash
make swag
```

Сгенерированные файлы \`docs/\` включены в репозиторий.

## Команды Makefile

```bash
make run
make build
make test
make swag
make migrate-up
make migrate-down
make docker-up
make smoke
```

## Ограничения и допущения

- формат даты: \`MM-YYYY\`;
- стоимость хранится как целое число;
- расчет total выполняется помесячно;
- \`end_date\` может отсутствовать;
- \`end_date\` не может быть раньше \`start_date\`;
- \`user_id\` должен быть валидным UUID.

## Репозиторий

```text
github.com/dvnpv/subscriptions-aggregator
```