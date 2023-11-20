# Use an official Golang runtime as a parent image
FROM golang:1.21

# Set the working directory to the project directory
WORKDIR /go/src/app

# Copy the local package files to the container's workspace
COPY . .

# Copy the .env file into the container
COPY .env .

# Build the application inside the container
RUN go build -o /go/bin/betrayal-bot ./cmd/betrayal-bot/

# Set the entry point for the application
CMD ["/go/bin/betrayal-bot"]
