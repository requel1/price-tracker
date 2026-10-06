# Price Tracker

Сервис для отслеживания цен на товары в интернет-магазинах. Периодически парсит цены и хранит историю изменений.

## Что делает

- Принимает товары (URL + название + сайт) через REST API
- Парсит текущую цену с сайта магазина
- Сохраняет историю цен в PostgreSQL
- Отдаёт историю через API
- Работает по расписанию: каждые 6 часов обновляет цены

## Стек

- **Go 1.22+** — язык
- **net/http** — HTTP-сервер (стандартная библиотека)
- **PostgreSQL 16** — база данных
- **pgx/v5** — драйвер PostgreSQL
- **colly/v2** — парсинг HTML
- **golang-migrate** — миграции БД
- **Docker / Docker Compose** — инфраструктура
- **slog** — структурированное логирование

## Архитектура

```
cmd/
├── api/       — HTTP API
└── worker/    — парсер по расписанию

internal/
├── config/      — конфиг из .env
├── model/       — модели данных
├── repository/  — работа с БД
├── service/     — бизнес-логика (парсинг)
├── parser/      — парсеры сайтов (Chitai-Gorod)
└── handler/     — HTTP-хендлеры

migrations/    — SQL-миграции
web/           — HTML с графиком
```

## Запуск

### Требования

- Go 1.22+
- Docker + Docker Compose
- `migrate` CLI

### Шаги

```bash
# 1. Клонировать
git clone https://github.com/requel1/price-tracker.git
cd price-tracker

# 2. Создать .env
cp .env.example .env

# 3. Поднять PostgreSQL
docker compose up -d postgres

# 4. Накатить миграции
make migrate-up

# 5. Запустить API (терминал 1)
go run ./cmd/api

# 6. Запустить воркер (терминал 2)
go run ./cmd/worker
```

API будет доступно на `http://localhost:8080`.

## API

Базовый URL: `http://localhost:8080`

### Создать товар

```
POST /api/products
Content-Type: application/json
```

Тело запроса:

```json
{
  "url": "https://www.chitai-gorod.ru/product/...",
  "name": "Выжить в качестве зены-героя. Том 3",
  "site": "chitai-gorod"
}
```

Ответ `201 Created`:

```json
{
  "id": 1,
  "url": "https://www.chitai-gorod.ru/product/...",
  "name": "Выжить в качестве зены-героя. Том 3",
  "site": "chitai-gorod",
  "created_at": "2026-10-06T20:25:15Z"
}
```

### Список товаров

```
GET /api/products
```

Ответ `200 OK`:

```json
[
  {
    "id": 1,
    "url": "...",
    "name": "...",
    "site": "chitai-gorod",
    "created_at": "..."
  }
]
```

### Получить товар

```
GET /api/products/{id}
```

Ответ `200 OK` — объект товара.

Ответ `404 Not Found` — если товара нет.

### Удалить товар

```
DELETE /api/products/{id}
```

Ответ `204 No Content` — успех.

Ответ `404 Not Found` — если товара нет.

### История цен

```
GET /api/products/{id}/prices
```

Ответ `200 OK`:

```json
[
  {
    "id": 1,
    "product_id": 1,
    "price": 1109,
    "currency": "RUB",
    "parsed_at": "2026-10-06T20:25:15Z"
  }
]
```

Сначала новые (сортировка по `parsed_at DESC`).

## Коды ответов

| Код | Когда |
|---|---|
| 200 | Успех |
| 201 | Товар создан |
| 204 | Товар удалён |
| 400 | Невалидный JSON / невалидный id |
| 404 | Товар не найден |
| 500 | Внутренняя ошибка |

## Примеры

```bash
# Создать товар
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.chitai-gorod.ru/product/vyzit-v-kacestve-zeny-geroa-tom-3-3183836",
    "name": "Выжить в качестве зены-героя. Том 3",
    "site": "chitai-gorod"
  }'

# Список
curl http://localhost:8080/api/products

# Один
curl http://localhost:8080/api/products/1

# История цен
curl http://localhost:8080/api/products/1/prices

# Удалить
curl -X DELETE http://localhost:8080/api/products/1
```

## Как работает воркер

1. Подключается к БД
2. Запускает `ParseAll` сразу
3. Дальше — каждые 6 часов

`ParseAll`:

- Берёт все товары из БД
- Запускает **5 горутин** (воркеров)
- Каждая берёт товар из канала `jobs`
- Парсит цену через `parser.GetParser(site)`
- Сохраняет в `prices`

**Почему 5 воркеров:** параллельно, но не перегружаем сайт. Больше — риск бана.

## Переменные окружения

Создай `.env` на основе `.env.example`:

```bash
cp .env.example .env
```

| Переменная | Описание | Пример |
|---|---|---|
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5433` |
| `DB_USER` | Пользователь БД | `postgres` |
| `DB_PASSWORD` | Пароль БД | — |
| `DB_NAME` | Имя базы | `price_tracker` |
| `DB_SSLMODE` | SSL-режим | `disable` |
| `API_PORT` | Порт API | `8080` |

## Makefile

```bash
make migrate-up                          # накатить миграции
make migrate-down                        # откатить одну
make migrate-create name=add_users       # создать миграцию
make migrate-version                     # текущая версия
make migrate-force version=N             # сбросить dirty
```

## Что дальше

- [ ] График цен на Chart.js
- [ ] Тесты (unit + integration)
- [ ] Middleware (CORS, recovery, logging)
- [ ] Деплой на VPS
- [ ] Поддержка других сайтов (Ozon, WB)