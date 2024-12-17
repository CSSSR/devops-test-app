# Stage 1: Build the application
FROM golang:1.23 AS builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Stage 2: Use Alpine as the runtime base
FROM alpine:latest

WORKDIR /app

# Create a non-root user and group "nonroot"
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot

# Copy the binary file from the builder stage
COPY --from=builder /app/app .

# Switch to the non-root user
USER nonroot:nonroot

# Run the application
CMD ["/app/app"]