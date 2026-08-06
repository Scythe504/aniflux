FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install make & build tools
RUN apk add --no-cache make git

# Copy dependency definitions
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Run prebuild and build worker targets
RUN make prebuild && make build-worker

FROM alpine:latest

WORKDIR /app

# Copy compiled worker binary & data
COPY --from=builder /app/bin/worker ./bin/worker
COPY --from=builder /app/Makefile ./Makefile
COPY --from=builder /app/data ./data

# Expose Stdin execution entrypoint
ENTRYPOINT ["/app/bin/worker"]
