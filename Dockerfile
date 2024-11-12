# Start from a small base image
FROM golang:1.23.3-alpine3.20 as builder

# Install build dependencies
RUN apk add --no-cache build-base

# Set the working directory
WORKDIR /app

# Copy only the necessary Go files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the Go application
RUN CGO_ENABLED=1 go build -o app ./cmd/ps2-spy/

# Create a minimal runtime image
FROM alpine:3.20.0

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/app .

ENTRYPOINT [ "./app" ]
