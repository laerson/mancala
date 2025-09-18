# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the notification service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o notifications ./cmd/notifications

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/notifications .

# Expose port
EXPOSE 50056

# Run the notification service
CMD ["./notifications"]