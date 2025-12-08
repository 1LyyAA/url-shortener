# этап сборки
FROM golang:1.22 AS builder
WORKDIR /app

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY ./src .
RUN CGO_ENABLED=0 GOOS=linux go build -o app main.go

# этап запуска
FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/app .
COPY --from=builder /app/index.html .

EXPOSE 8080

CMD ["./app"]
