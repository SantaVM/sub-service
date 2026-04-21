# Этап 1: Сборка
FROM golang:1.25.9-alpine3.22 AS builder

WORKDIR /app

# Установка необходимых пакетов
RUN apk add --no-cache git

# Копирование go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o /sub-service ./cmd/server/main.go

# Этап 2: Финальный образ
FROM alpine:3.22

WORKDIR /app

# Установка ca-certificates для HTTPS
RUN apk --no-cache add ca-certificates

# Копирование бинарного файла
COPY --from=builder /sub-service /app/sub-service
COPY --from=builder /app/docs /app/docs

RUN adduser -D appuser
USER appuser

# Expose port
EXPOSE 8080

# Запуск приложения
CMD ["/app/sub-service"]
