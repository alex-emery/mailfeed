# Use the official Go image as the base image
FROM golang:1.21 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download and install the Go dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN make build 

FROM golang:1.21

WORKDIR /app 

COPY --from=builder /app/bin/mailfeed /bin/mailfeed
# COPY /bin/mailfeed /usr/bin/mailfeed

CMD ["mailfeed", "--db=/data/mailfeed.db", "--host=mailfeed.xyz"]