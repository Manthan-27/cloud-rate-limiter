# Step 1: Build the Go app
FROM golang:1.23.4-alpine AS builder

# Install git for module fetching
RUN apk add --no-cache git

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum first to leverage caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

# Build the Go app, output binary named "app"
RUN go build -o app ./cmd/server

# Step 2: Use a minimal image to run the app
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/app .

# Expose server port
EXPOSE 8080

# Run the app
CMD ["./app"]
