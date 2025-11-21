# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
COPY ../common-service/go.mod ../common-service/go.sum ../common-service/

# Download dependencies
RUN go mod download

# Copy source code
COPY . .
COPY ../common-service ../common-service

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway ./cmd/api-gateway

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary and config
COPY --from=builder /app/api-gateway .
COPY --from=builder /app/config.yaml .

EXPOSE 8080

CMD ["./api-gateway"]

