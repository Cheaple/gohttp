# Use the official Go image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the local source files to the container
COPY . .

# Build the Go application
RUN go build -o proxy/main ./proxy/proxy.go

# Expose a port (if your Go program listens on a specific port)
EXPOSE 8888

# Command to run the Go application
CMD ["./proxy/main", "-p=8888"]
