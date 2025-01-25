# Use the official Go image as the base image
FROM golang:1.23.4-bullseye

# Install essential build tools
RUN apt-get update && apt-get install -y make curl

# Install migrate CLI (version 4.18.1) directly from GitHub release
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xvz \
    && mv migrate /usr/local/bin/migrate \
    && chmod +x /usr/local/bin/migrate

# Set the working directory
WORKDIR /app

# Copy Go module files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Install development tools
RUN go install github.com/air-verse/air@latest \
    && go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger documentation
RUN make gen-docs

# Start the application with air for hot-reloading
CMD ["air", "-c", ".air.toml"]