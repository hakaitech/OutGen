FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /OutGen

# We want to populate the module cache based on the go.{mod,sum} files.
COPY OutGen/go.mod ./
COPY OutGen/go.sum ./

RUN go mod download
COPY OutGen/ /OutGen

# Build the Go app
RUN go build -o OG

# Run the binary program produced by `go install`
CMD ["./OG"]
