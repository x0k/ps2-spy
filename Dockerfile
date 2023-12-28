# Start from a small base image
FROM golang:1.21.5-alpine3.19 as builder

# Set the working directory
WORKDIR /app

# Copy only the necessary Go files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the Go application
RUN go build -o app ./cmd/ps2-feed-ds

# Create a minimal runtime image
FROM alpine:3.19.0

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/app .

ENTRYPOINT [ "./app" ]
