# wallet-api

REST API для личных финансов на Go 1.22: регистрация и логин по JWT, кошельки, транзакции, атомарное обновление баланса в PostgreSQL.

## Стек

- Go 1.22
- Chi router `github.com/go-chi/chi/v5`
- PostgreSQL 16 + `pgx/v5` / `pgxpool`
- JWT `github.com/golang-jwt/jwt/v5`
- Миграции `golang-migrate/migrate/v4`
- bcrypt, UUID, decimal
- Unit-тесты `testify` + `testify/mock`
- Docker Compose: app + postgres
- `.env` через `cleanenv`

## Быстрый старт

```bash
cp .env.example .env
docker compose up --build
```

API будет доступен на `http://localhost:8080`.

Проверка здоровья:

```bash
curl http://localhost:8080/health
```

## Локальный запуск без Docker для приложения

PostgreSQL можно поднять через Compose, а приложение запустить локально:

```bash
cp .env.example .env
docker compose up -d postgres
export DATABASE_URL='postgres://wallet:wallet@localhost:5432/wallet_api?sslmode=disable'
export JWT_SECRET='change-me-super-secret-key'
go run ./cmd/server
```

Миграции запускаются автоматически при старте сервера.

## Тесты и проверки

```bash
go test ./...
go vet ./...
golangci-lint run
```

## Переменные окружения

| Переменная | Описание | Пример |
|---|---|---|
| `ENV` | окружение | `local` |
| `HTTP_PORT` | порт HTTP сервера | `8080` |
| `DATABASE_URL` | строка подключения PostgreSQL | `postgres://wallet:wallet@postgres:5432/wallet_api?sslmode=disable` |
| `JWT_SECRET` | секрет для подписи JWT | `change-me-super-secret-key` |
| `JWT_TTL` | срок жизни JWT | `24h` |
| `MIGRATIONS_PATH` | путь к миграциям migrate | `file://migrations` |
| `POSTGRES_DB` | имя БД для Docker Compose | `wallet_api` |
| `POSTGRES_USER` | пользователь БД | `wallet` |
| `POSTGRES_PASSWORD` | пароль БД | `wallet` |

## Эндпоинты

| Метод | Путь | JWT | Описание |
|---|---|---:|---|
| `POST` | `/api/v1/auth/register` | нет | регистрация |
| `POST` | `/api/v1/auth/login` | нет | логин |
| `GET` | `/api/v1/wallets` | да | список кошельков пользователя |
| `POST` | `/api/v1/wallets` | да | создание кошелька |
| `GET` | `/api/v1/wallets/{id}` | да | один кошелёк |
| `DELETE` | `/api/v1/wallets/{id}` | да | удалить кошелёк, только если баланс равен нулю |
| `GET` | `/api/v1/wallets/{id}/transactions?page=1&limit=20` | да | транзакции кошелька с пагинацией |
| `POST` | `/api/v1/wallets/{id}/transactions` | да | создать транзакцию |
| `GET` | `/api/v1/transactions?from=2024-01-01&to=2024-12-31` | да | все транзакции пользователя за период |

## Примеры curl

### Регистрация

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"password123"}'
```

Ответ:

```json
{
  "token": "<jwt>",
  "user": {
    "id": "<uuid>",
    "email": "user@example.com",
    "created_at": "2026-06-16T12:00:00Z"
  }
}
```

Сохраните токен:

```bash
TOKEN='<jwt>'
```

### Логин

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"password123"}'
```

### Создать кошелёк

```bash
curl -s -X POST http://localhost:8080/api/v1/wallets \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Main","currency":"RUB"}'
```

### Получить кошельки

```bash
curl -s http://localhost:8080/api/v1/wallets \
  -H "Authorization: Bearer $TOKEN"
```

### Получить один кошелёк

```bash
WALLET_ID='<wallet_uuid>'
curl -s http://localhost:8080/api/v1/wallets/$WALLET_ID \
  -H "Authorization: Bearer $TOKEN"
```

### Создать доход

```bash
curl -s -X POST http://localhost:8080/api/v1/wallets/$WALLET_ID/transactions \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"type":"income","amount":"1000.00","description":"salary"}'
```

### Создать расход

```bash
curl -s -X POST http://localhost:8080/api/v1/wallets/$WALLET_ID/transactions \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"type":"expense","amount":"150.50","description":"groceries"}'
```

Если денег недостаточно, API вернёт:

```json
{ "error": "insufficient funds" }
```

### История транзакций кошелька

```bash
curl -s "http://localhost:8080/api/v1/wallets/$WALLET_ID/transactions?page=1&limit=20" \
  -H "Authorization: Bearer $TOKEN"
```

### Все транзакции пользователя за период

```bash
curl -s "http://localhost:8080/api/v1/transactions?from=2024-01-01&to=2024-12-31" \
  -H "Authorization: Bearer $TOKEN"
```

### Удалить кошелёк

Удаление возможно только при балансе `0.00`.

```bash
curl -i -X DELETE http://localhost:8080/api/v1/wallets/$WALLET_ID \
  -H "Authorization: Bearer $TOKEN"
```

## Архитектура

- `handler`: парсинг HTTP-запросов, вызов сервисов, JSON-ответы.
- `service`: бизнес-логика, проверки баланса, проверка владельца ресурсов.
- `repository`: SQL через pgx без ORM.
- `middleware`: JWT-аутентификация и логирование запросов.
- `migrations`: схема PostgreSQL с индексами.

Баланс кошелька обновляется атомарно: сервис открывает транзакцию БД, блокирует кошелёк через `FOR UPDATE`, проверяет бизнес-правила, обновляет баланс и создаёт запись транзакции.
