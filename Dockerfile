# Use an official Go runtime as a parent image
FROM golang:1.22 as builder

# Set the working directory inside the container
WORKDIR /go/src/app

# Copy the local package files to the container's workspace.
COPY . .

# Build your program for Linux.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o bundler-node ./cmd/bundler-node

# Use a Docker multi-stage build to create a lean production image.
# https://docs.docker.com/develop/develop-images/multistage-build/
FROM alpine:latest  

# Add ca-certificates in case you need HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary and sample config from the builder stage.
COPY --from=builder /go/src/app/bundler-node .
COPY --from=builder /go/src/app/etc/bundler-node.conf.sample ./bundler-node.conf

# Set the default config file environment variable
ENV BUNDLER_NODE_CONFIG="./bundler-node.conf"

# Use ENTRYPOINT to specify the binary
ENTRYPOINT ["./bundler-node"]

# Use CMD to specify the default arguments to the ENTRYPOINT
CMD ["--config", "./bundler-node.conf"]