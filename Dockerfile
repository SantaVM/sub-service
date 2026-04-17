# Этап 1: Сборка
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Установка необходимых пакетов
RUN apk add --no-cache git

# Копирование go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Установка swag для генерации документации
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Генерация Swagger документации
RUN swag init -g cmd/server/main.go -o docs

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o /sub-service ./cmd/server/main.go

# Этап 2: Финальный образ
FROM alpine:latest

WORKDIR /app

# Установка ca-certificates для HTTPS
RUN apk --no-cache add ca-certificates

# Копирование бинарного файла
COPY --from=builder /sub-service /app/sub-service
COPY --from=builder /app/docs /app/docs

# Expose port
EXPOSE 8080

# Запуск приложения
CMD ["/app/sub-service"]
