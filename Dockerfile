# Use an official Golang runtime as a parent image
FROM golang:1.22

# Set the working directory to /app
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
#
# Download all dependencies
RUN go mod download

#Load Environment Variables from .env file 
ARG ENV_FILE
ENV ENV_FILE=${ENV_FILE}
COPY $ENV_FILE .env

# Copy the source code from the current directory and subdirectories to the working directory inside the container
COPY . .


# Build the application
RUN go build -o ./bin/main /app/cmd/betrayal-bot/

# Run the binary program produced by `go install`
CMD ["./bin/main"]
