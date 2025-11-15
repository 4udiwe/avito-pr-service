# Step 1: Modules caching
FROM golang:1.24-alpine3.22 AS modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Builder
FROM golang:1.24-alpine3.22 AS builder
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/avito-pr ./cmd/main.go

# Step 3: Final
FROM alpine:3.22
COPY --from=builder /app/config /config
COPY --from=builder /app/avito-pr /app/avito-pr

WORKDIR /app
EXPOSE 8080
CMD ["/app/avito-pr"]
