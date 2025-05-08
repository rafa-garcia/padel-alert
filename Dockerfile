# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy only the go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a static binary (no external dependencies)
# CGO_ENABLED=0: Disables C-Go, making a fully static binary
# -ldflags="-s -w": Strips debug information to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /padel-alert ./cmd/padel-alert

# Final stage
FROM scratch

# Copy SSL certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /padel-alert /padel-alert

EXPOSE ${PORT:-8080}

# Set environment variable defaults
ENV PORT=8080
ENV LOG_LEVEL=info

ENTRYPOINT ["/padel-alert"]
