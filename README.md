# Subscription Service

REST-сервис для агрегации данных об онлайн подписках пользователей.

## Возможности

- CRUDL операции над подписками
- Валидация входящих данных с понятными ошибками
- Подсчет суммарной стоимости подписок за выбранный период
- Фильтрация по ID пользователя и названию сервиса
- PostgreSQL с автоматическими миграциями
- Swagger документация
- Docker Compose для простого запуска
- Генерация UUIDv7 для создания id пользователя
- Юнит-тесты для основной логики хэндлеров
- Интеграционный тест для операции создания Подписки с запуском БД в Docker с использованием testcontainers
- Тесты основных ограничений, заданных на уровне БД, с запуском БД в Docker с использованием testcontainers
- Миграции для инициализации БД

## Бизнес-ограничения

- Если время окончания подписки не задано, то считается, что подписка длится вечно
- Один пользователь не может иметь две подписки на один и тот же сервис с пересекающимися датами (проверяется на уровне БД)

## Структура проекта

```
├───cmd/server/main.go        - Точка входа
├───docs                      - Документы Swagger OPENAPI
├───internal
│    ├───app                   - Cборка приложения (DI)
│    ├───handler               - HTTP обработчики
│    │   └───middleware        - HTTP фильтры
│    ├───infrastructure
│    │   ├───config            - Конфигурация
│    │   ├───database          - Работа с БД и миграции
│    │   │   └───migrations    - Файлы миграций
│    │   ├───logger            - Настройка структурированных логов
│    │   └───validator         - Настройка валидатора входящих данных
│    ├───model                 - Модели данных и ДТО
│    ├───repository            - Слой доступа к БД
│    ├───service               - Бизнес-логика
│    └───testutil              - Вспомогательные файлы для запуска тестов БД
├── docker-compose.yml         - Docker конфигурация
├── Dockerfile                 - Docker образ
└── .env.local                 - Переменные окружения
```

## Быстрый старт

### Запуск с помощью Docker Compose

1. Убедитесь, что в системе установлены:
    - Docker и Docker Compose
    - Git

2. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/SantaVM/sub-service.git
   ```

3. Запустите сервис:
   ```bash
   docker-compose up -d
   ```

4. Сервис будет доступен на `http://localhost:8080`
5. Swagger UI на `http://localhost:8080/swagger/index.html`

### Локальная разработка

1. Установите зависимости:
   ```bash
   go mod download
   ```

2. Запустите PostgreSQL (например, через Docker):
   ```bash
   docker run -d -p 15432:5432 -e POSTGRES_DB=subservice -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=example postgres:17-alpine
   ```

3. Установите swag CLI:
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

4. Установите приложение go-task (команда для linux / Windows):
   ```bash
   npm install -g @go-task/cli
   ```

   или (только для Windows):
   ```ps
   winget install Task.Task
   ```

5. Запустите сервис:
   ```bash
   task
   ```

## API Endpoints

### Подписки

- `POST /api/v1/subscriptions` - Создание подписки
- `GET /api/v1/subscriptions` - Список подписок
- `GET /api/v1/subscriptions/{id}` - Получение подписки по ID
- `PUT /api/v1/subscriptions/{id}` - Обновление подписки
- `DELETE /api/v1/subscriptions/{id}` - Удаление подписки
- `GET /api/v1/subscriptions/total` - Подсчет суммарной стоимости
- `GET /api/v1/uuid` - Генерирует UUIDv7 для применения при создании подписки в БД в качестве ID пользователя

### Пример запроса на создание подписки

```json
{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025",
  "end_date": "07-2025"
}
```

## Конфигурация

Все настройки вынесены в `.env.local` файл:

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| PORT | Порт сервера | 8080 |
| DB_HOST | Хост базы данных | localhost |
| DB_PORT | Порт базы данных | 5432 |
| DB_NAME | Имя базы данных | subservice |
| DB_USER | Пользователь базы данных | postgres |
| DB_PASSWORD | Пароль базы данных | postgres |

## Технологии

- **Go** - основной язык
- **chi** - HTTP роутер
- **PostgreSQL** - база данных
- **goose** - миграции БД
- **swaggo** - Swagger документация
- **Docker & Docker Compose** - контейнеризация
