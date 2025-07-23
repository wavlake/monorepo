FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code explicitly
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
COPY tests/ tests/

# Build both API server and file server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o fileserver ./cmd/fileserver

FROM alpine:3.19

# Install ffmpeg and ca-certificates for audio processing
RUN apk add --no-cache ffmpeg ca-certificates

# Copy built binaries
COPY --from=builder /app/api /api
COPY --from=builder /app/fileserver /fileserver

# Set environment variables
ENV PORT=8080

# Expose ports
EXPOSE 8080
EXPOSE 8081

# Default to running the API server
ENTRYPOINT ["/api"]

# To run fileserver instead, override with: docker run ... /fileserver