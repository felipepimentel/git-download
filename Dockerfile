# Build stage
FROM golang:1.18-alpine AS builder

WORKDIR /app
COPY . .

# Build the binary with version information
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o git-download ./cmd/git-download

# Final stage
FROM alpine:latest

# Add ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/git-download .

# Create a directory for metadata
RUN mkdir -p /data
VOLUME /data
WORKDIR /data

ENTRYPOINT ["/root/git-download"] 