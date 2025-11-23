FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

COPY --from=builder /app/migrations ./migrations

COPY --from=builder /go/bin/goose /usr/local/bin/goose

RUN ls -la /root/migrations/

# Экспонируем порт
EXPOSE 8080

CMD ["./main"]