# First stage: Get Golang image from DockerHub.
FROM golang:1.21.1 AS backend-builder

# Label this container.
LABEL appname="Betrayal Web App"
LABEL author="Alex McCune <alexmccune1224@gmail.com>"
LABEL description="Simple Go Echo backend + React frontend for Betrayal game."

# Set our working directory for this stage.
WORKDIR /backendcompile

# Copy all of our files.
COPY . .

# Get and install all dependencies.
RUN CGO_ENABLED=0 go build -o web ./cmd/web/.

# Next stage: Build our frontend application.
FROM node:20 AS frontend-builder

# Set our working directory for this stage.
WORKDIR /frontendcompile

# Copy lockfiles and dependencies.
COPY ./www/package.json ./www/yarn.lock ./

# Install our dependencies.
RUN yarn

# Copy our installed 'node_modules' and everything else.
COPY ./www .

# Build our application.
RUN yarn build

# Last stage: discard everything except our executables.
FROM alpine:latest AS prod

# Set our next working directory.
WORKDIR /build

# Create directory for our React application to live.
RUN mkdir -p /www/build

# Copy our executable and our built React application.
COPY --from=backend-builder /backendcompile/web .
COPY --from=frontend-builder /frontendcompile/build ./www/build

# Declare entrypoints and activation commands.
EXPOSE 8080
ENTRYPOINT ["./web"]
