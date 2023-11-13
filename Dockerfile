# Use the official Go image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the local source files to the container
COPY . .

# Build the Go application
RUN go build -o ./server/main ./server/server.go

# Expose a port (if your Go program listens on a specific port)
EXPOSE 8080

# Command to run the Go application
CMD ["./server/main", "-p=8080"]
