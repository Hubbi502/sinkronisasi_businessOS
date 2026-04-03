# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build both applications
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api_bin ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/worker_bin ./cmd/worker

# Final stage
FROM alpine:latest

# Install ca-certificates for external API calls (e.g., Google Sheets)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binaries
COPY --from=builder /app/api_bin /app/api_bin
COPY --from=builder /app/worker_bin /app/worker_bin

# Expose API port
EXPOSE 8080

# Docker compose will override the command depending on the service
CMD ["/app/api_bin"]
