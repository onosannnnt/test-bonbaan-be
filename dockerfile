# Use an official Golang runtime as a parent image
FROM golang:1.17-alpine

# Set the working directory in the container
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . .

# Download and install any needed dependencies
RUN go mod download

# Build the Go app
RUN go build -o main .

# Make port 80 available to the world outside this container
EXPOSE 80

# Run the executable
CMD ["./main"]  