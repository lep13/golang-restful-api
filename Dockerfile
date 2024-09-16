# Step 1: Use an official Golang image as the base (alpine variant is lightweight)
FROM golang:1.22-alpine AS builder

# Step 2: Set the working directory in the container
WORKDIR /app

# Step 3: Copy go.mod and go.sum files to download dependencies
COPY . .
COPY .env .
COPY go.mod go.sum ./
RUN go mod download

# Step 5: Build the Go app binary
RUN go build -o main .

# Step 6: Use a smaller base image to run the Go app (optional)
FROM alpine:latest
WORKDIR /root/

# Step 7: Copy the built binary from the builder image
COPY --from=builder /app/main .


# Step 9: Expose the necessary port (5000 as per your setup)
EXPOSE 5000

# Step 10: Command to run the application
CMD ["./main"]
