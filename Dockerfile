FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/test_file_service ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/test_file_service .
COPY .env .

EXPOSE 8000

CMD ["./test_file_service"]