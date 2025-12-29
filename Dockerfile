# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 creates a fully static binary
# -ldflags="-w -s" strips debug info to reduce size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o worker \
    ./cmd/worker

# Certificates stage - get CA certificates and timezone data
FROM alpine:latest AS certs
RUN apk --no-cache add ca-certificates tzdata

# Runtime stage - minimal scratch image
FROM scratch

# Copy CA certificates from certs stage (needed for HTTPS requests)
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy only São Paulo timezone (instead of all timezones)
COPY --from=certs /usr/share/zoneinfo/America/Sao_Paulo /usr/share/zoneinfo/America/Sao_Paulo

# Set timezone to São Paulo
ENV TZ=America/Sao_Paulo

# Copy the static binary from builder
COPY --from=builder /build/worker /worker

# Execute the worker
# Note: Running as root in scratch (no user management available)
# This is acceptable for Cloud Run Jobs with proper IAM
ENTRYPOINT ["/worker"]
