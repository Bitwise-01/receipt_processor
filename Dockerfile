# Dockerfile
FROM golang:1.23 AS builder

WORKDIR /app

# Copy go.mod and go.sum files to download dependencies.
COPY go.mod go.sum ./
COPY config /config/
RUN go mod download

# Copy the entire project.
COPY . .

# Build the binary. We use CGO_ENABLED=0 for a statically-linked binary.
RUN CGO_ENABLED=0 go build -o receipt_processor ./cmd/receipt_processor

FROM alpine:latest

# Install certificates (if your app makes HTTPS calls, etc.).
RUN apk --no-cache add ca-certificates

# Set the working directory.
WORKDIR /root/

# Copy the binary from the builder stage.
COPY --from=builder /app/receipt_processor .

# **Copy the config folder from the builder stage to the final container.**
COPY --from=builder /app/config ./config

# Expose the port (ensure this matches your config, default is 8080).
EXPOSE 8080

# Run the binary.
CMD ["./receipt_processor"]
