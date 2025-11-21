# Build stage
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Run stage
FROM alpine:latest
WORKDIR /root/

# Install CA certificates for SSL database connections
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy templates directory
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/internal ./internal

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
