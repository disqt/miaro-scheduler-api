# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o miaro-scheduler-api .

# Final stage
FROM alpine:latest

# Install ca-certificates and tzdata for HTTPS and timezone support
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/miaro-scheduler-api .

# Copy templates
COPY --from=builder /app/templates ./templates

# Expose port
EXPOSE 8081

# Set environment variables with defaults
ENV PORT=8081
ENV TIMEZONE=Europe/Paris
ENV ENABLE_CORS=false

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/health || exit 1

# Run the application
CMD ["./miaro-scheduler-api"]
