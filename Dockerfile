# -------- Stage 1: Build the Go app --------
FROM golang:1.21 AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum first (better for caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o main

# -------- Stage 2: Create a small runtime image --------
FROM alpine:latest

# Install required certificates for HTTPS (if needed)
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Expose the application's port (change if needed)
EXPOSE 3000

# Run the application
CMD ["./main"]
