FROM golang:1.24-alpine AS builder

# Install security updates
RUN apk update && apk upgrade && apk add --no-cache ca-certificates

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with security flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static'" \
    -a -installsuffix cgo \
    -o server ./cmd/server

# Final stage - distroless for minimal attack surface
FROM gcr.io/distroless/static:nonroot

# Copy the binary from builder (static files are embedded in the binary)
COPY --from=builder /app/server /server

# Use non-root user (65532 is the nonroot user in distroless)
USER nonroot:nonroot

# Expose port (informational)
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/server"]