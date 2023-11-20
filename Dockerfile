# Use the official Golang image as the base image
FROM golang:1.21

# Set the working directory inside the container
WORKDIR /app

# Copy the Go application source code and .env file to the container
COPY ./cmd/betrayal-bot /app
COPY ./internal /app/internal
COPY .env /app

# Build the Go application
RUN go build -o betrayal-bot .


# Command to run the application
CMD ["./betrayal-bot"]
