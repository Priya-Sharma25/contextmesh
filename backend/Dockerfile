# --- Stage 1: Build compilation ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install standard build tools
RUN apk add --no-cache git gcc musl-dev

# Cache dependency layers
COPY go.mod go.sum ./
RUN go mod download

# Copy source trees
COPY . .

# Compile optimized static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o contextmesh-server cmd/server/main.go

# --- Stage 2: Distroless-like lightweight runner ---
FROM alpine:3.19

WORKDIR /root/

# Copy compiled static binary
COPY --from=builder /app/contextmesh-server .

EXPOSE 8080

# Configure production microservices options
ENV PORT=8080
ENV GODEBUG=netdns=go

ENTRYPOINT ["./contextmesh-server"]
