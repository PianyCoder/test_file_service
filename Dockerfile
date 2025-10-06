FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o test_file_service ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/test_file_service .
COPY .env .

EXPOSE 8080

CMD ["./test_file_service"]