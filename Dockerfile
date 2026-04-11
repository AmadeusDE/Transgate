# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o transgate ./cmd/transgate

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies (if any)
RUN apk add --no-cache ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/transgate .

# Copy the example configuration
COPY --from=builder /app/config.example.json ./config.example.json

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./transgate"]
