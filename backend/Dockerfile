# Build stage
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application — TARGETOS/TARGETARCH set by buildx for multi-platform
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o pageturner .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS and bash for startup script
RUN apk add --no-cache ca-certificates bash

# Copy binary from builder
COPY --from=builder /app/pageturner .
COPY --from=builder /app/config-docker.yml ./config.yml
COPY --from=builder /app/migrations ./migrations

# Copy startup script
COPY start.sh /start.sh
RUN chmod +x /start.sh

# Expose port
EXPOSE 8080

# Run the application with startup script
CMD ["/start.sh"]