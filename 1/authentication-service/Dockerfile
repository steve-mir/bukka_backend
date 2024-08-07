# Use a smaller base image like Alpine
FROM golang:1.22.3-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy only necessary files to the container's workspace
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Run ls command to list files in the /app directory
RUN ls -l /app

# Build the application
# RUN CGO_ENABLED=0 go build -o main ./cmd/api
RUN CGO_ENABLED=0 go build -o main ./main.go

# Use a lightweight base image for the final runtime image
FROM alpine:latest

# Copy the binary from the builder stage to the final image
COPY --from=builder /app/main /app/main


# # Expose the port
# ARG GRPC_PORT
# EXPOSE 6000 $GRPC_PORT
# # EXPOSE 7001 6000

# Run the executable
CMD ["/app/authentication-service"]
# ENTRYPOINT ["/app/authentication-service"]