# FROM alpine:latest

# RUN mkdir /app

# COPY authApp /app

# CMD [ "/app/authApp"]

# ***
# FROM alpine:latest

# RUN mkdir /app

# COPY . /app

# CMD [ "/app/authentication-service" ]

# ***

# Use a minimal base image like Alpine with Go pre-installed
FROM golang:1.22.3-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to cache dependencies
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if go.mod and go.sum are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o authentication-service ./cmd/authentication-service

# Start a new stage from scratch
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the builder stage
COPY --from=builder /app/authentication-service .

# Provide executable permissions
RUN chmod +x ./authentication-service

# Command to run the executable
CMD ["./authentication-service"]
