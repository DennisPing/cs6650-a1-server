# https://hub.docker.com/_/golang
FROM golang:1.19-buster AS builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy server code to container.
COPY cmd/ ./cmd
COPY log/ ./log
COPY metrics/ ./metrics
COPY models/ ./models
COPY server/ ./server

# Build the binary.
RUN go build -v -o httpserver ./cmd

# Use the official Debian slim image for a lean production container.
FROM debian:buster-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/httpserver /app/httpserver

# Run the web service on container startup.
CMD ["/app/httpserver"]
