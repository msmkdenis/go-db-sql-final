# Используем официальный образ Golang для создания сборочного артефакта.
FROM golang:1.21 AS builder

# Создаем и переходим в директорию приложения.
WORKDIR /app

# Копируем go.mod и go.sum для загрузки зависимостей.
COPY go.mod go.sum ./

# Загружаем все зависимости. Зависимости будут кэшированы, если файлы go.mod и go.sum не были изменены.
RUN go mod download

# Копируем исходный код из текущего каталога в рабочий каталог внутри контейнера.
COPY . .

# Собираем Go-приложение.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Используем образ alpine для создания небольшого конечного образа.
FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Копируем предварительно собранный исполняемый файл из предыдущего этапа.
COPY --from=builder /app/app /app/app

# Команда для запуска исполняемого файла.
CMD ["/app/app"]