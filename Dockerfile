# Builder phase
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
# Download dependencies and build the application
RUN go mod download
RUN go build -o log-analyzer ./cmd/main.go

# Final phase
FROM alpine:latest
WORKDIR /root/
# Just copy the built binary and config files from the builder phase
COPY --from=builder /app/log-analyzer .
COPY --from=builder /app/config/rules.yaml ./config/

# Allocate for log storage
VOLUME ["/logs"]

# Run the application
CMD ["./log-analyzer"]