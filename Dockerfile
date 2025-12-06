# Use an official Golang runtime as a parent image
FROM golang:1.23

# Set the working directory to /app
WORKDIR /app

# Install templ CLI
RUN go install github.com/a-h/templ/cmd/templ@latest

# Install Tailwind CSS standalone CLI
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.1/tailwindcss-linux-x64 \
    && chmod +x tailwindcss-linux-x64 \
    && mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Load Environment Variables from .env file 
ARG ENV_FILE
ENV ENV_FILE=${ENV_FILE}
COPY $ENV_FILE .env

# Copy the source code from the current directory and subdirectories to the working directory inside the container
COPY . .

# Generate templ templates
RUN templ generate

# Build Tailwind CSS
RUN tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --minify

# Build the application
RUN go build -o ./bin/main /app/cmd/betrayal-bot/

# Expose web admin port (default 8080)
EXPOSE 8080

# Run the binary program produced by `go install`
CMD ["./bin/main"]
