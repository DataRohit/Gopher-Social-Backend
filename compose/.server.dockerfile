# Use the official Go image as the base image
FROM golang:1.23.4-bullseye AS build-stage

# Install make
RUN apt-get update && apt-get install -y make

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Download and cache the dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Install other dependencies
RUN go install github.com/air-verse/air@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger documentation
RUN make gen-docs

# Run the air server
CMD ["air", "-c", ".air.toml"]