# Step 1: Builder
FROM golang:1.24-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/avito-pr ./cmd/main.go

# Step 2: Final
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/avito-pr /app/avito-pr
COPY --from=builder /app/config/config.yaml /app/config/config.yaml
COPY --from=builder /app/internal/database/migrations /app/database/migrations

WORKDIR /app
EXPOSE 8080
CMD ["/app/avito-pr"]
