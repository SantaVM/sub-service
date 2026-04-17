# Subscription Service

REST-сервис для агрегации данных об онлайн подписках пользователей.

## Возможности

- CRUDL операции над подписками
- Подсчет суммарной стоимости подписок за выбранный период
- Фильтрация по ID пользователя и названию сервиса
- PostgreSQL с автоматическими миграциями
- Swagger документация
- Docker Compose для простого запуска

## Структура проекта

```
├── cmd/server/main.go          - Точка входа
├── internal/
│   ├── config/                 - Конфигурация
│   ├── handler/                - HTTP обработчики
│   ├── model/                  - Модели данных
│   ├── repository/             - Слой доступа к БД
│   └── service/                - Бизнес-логика
├── pkg/database/               - Работа с БД и миграции
├── docker-compose.yml          - Docker конфигурация
├── Dockerfile                  - Docker образ
└── .env                        - Переменные окружения
```

## Быстрый старт

### Запуск с помощью Docker Compose

1. Скопируйте `.env` файл и настройте переменные окружения:
   ```bash
   cp .env.example .env
   ```

2. Запустите сервис:
   ```bash
   docker-compose up -d
   ```

3. Сервис будет доступен на `http://localhost:8080`
4. Swagger UI на `http://localhost:8081`

### Локальная разработка

1. Установите зависимости:
   ```bash
   go mod download
   ```

2. Запустите PostgreSQL (например, через Docker):
   ```bash
   docker run -d -p 5432:5432 -e POSTGRES_DB=subservice -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres postgres:17-alpine
   ```

3. Установите swag CLI:
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

4. Сгенерируйте Swagger документацию:
   ```bash
   swag init -g cmd/server/main.go -o docs
   ```

5. Запустите сервис:
   ```bash
   go run cmd/server/main.go
   ```

## API Endpoints

### Подписки

- `POST /api/v1/subscriptions` - Создание подписки
- `GET /api/v1/subscriptions` - Список подписок
- `GET /api/v1/subscriptions/{id}` - Получение подписки по ID
- `PUT /api/v1/subscriptions/{id}` - Обновление подписки
- `DELETE /api/v1/subscriptions/{id}` - Удаление подписки
- `GET /api/v1/subscriptions/total` - Подсчет суммарной стоимости

### Пример запроса на создание подписки

```json
{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025"
}
```

### Пример запроса на подсчет стоимости

```
GET /api/v1/subscriptions/total?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&start_month=1&start_year=2025&end_month=12&end_year=2025
```

## Конфигурация

Все настройки вынесены в `.env` файл:

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| PORT | Порт сервера | 8080 |
| DB_HOST | Хост базы данных | localhost |
| DB_PORT | Порт базы данных | 5432 |
| DB_NAME | Имя базы данных | subservice |
| DB_USER | Пользователь базы данных | postgres |
| DB_PASSWORD | Пароль базы данных | postgres |
| LOG_LEVEL | Уровень логирования | info |

## Технологии

- **Go** - основной язык
- **chi** - HTTP роутер
- **PostgreSQL** - база данных
- **golang-migrate** - миграции БД
- **swaggo** - Swagger документация
- **Docker & Docker Compose** - контейнеризация
