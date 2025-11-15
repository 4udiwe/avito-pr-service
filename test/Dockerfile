# Step 1: Modules caching
FROM golang:1.24-alpine3.21 AS modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.24-alpine3.21
COPY --from=modules /go/pkg /go/pkg
COPY --from=modules /go/pkg/mod /go/pkg/mod

COPY . /app
WORKDIR /app

CMD ["go", "test", "-v", "-tags=integration", "./test/integration_test_deactivate_team/..."]