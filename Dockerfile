# Build the static SvelteKit application before compiling Go, so the binary
# embeds the exact production UI output.
FROM node:22-bookworm AS frontend-build
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./frontend/
RUN npm --prefix frontend ci
COPY . .
RUN npm --prefix frontend run build

# Use an official Golang runtime as a parent image
FROM golang:1.25

# Set the working directory to /app
WORKDIR /app

# Install templ CLI
RUN go install github.com/a-h/templ/cmd/templ@v0.3.960

# Install Tailwind CSS v4 standalone CLI
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.1.2/tailwindcss-linux-x64 \
    && chmod +x tailwindcss-linux-x64 \
    && mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code from the current directory and subdirectories to the working directory inside the container
COPY . .
COPY --from=frontend-build /app/internal/web/ui/dist ./internal/web/ui/dist

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
