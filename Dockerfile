# Accept the Go version for the image to be set as a build argument.
# Default to Go 1.11
ARG GO_VERSION=1.11

# First stage: build the executable.
FROM golang:${GO_VERSION}-alpine AS builder

# Create the user and group files that will be used in the running container to
# run the process an unprivileged user.
RUN mkdir /user && \
    echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
    echo 'nobody:x:65534:' > /user/group

# Install the Certificate-Authority certificates for the app to be able to make
# calls to HTTPS endpoints.
# Git is required for fetching the dependencies.
RUN apk add --no-cache ca-certificates git make

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /go/src/github.com/spothero/segment-proxy

# Fetch dependencies first; they are less susceptible to change on every build
# and will therefore be cached for speeding up the next build
# Copy necessary build files
RUN mkdir -p /go/src/github.com/spothero/segment-proxy
COPY Makefile .
COPY Gopkg.toml .
COPY Gopkg.lock .
COPY vendor/ ./vendor
COPY main.go .
COPY pkg/ ./pkg

# Build the executable to `/app`. Mark the build as statically linked.
RUN make build

# Final stage: the running container.
FROM alpine:3.8 AS final

# Import the user and group files from the first stage.
COPY --from=builder /user/group /user/passwd /etc/

# Import the Certificate-Authority certificates for enabling HTTPS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Import the compiled executable from the first stage.
COPY --from=builder /go/src/github.com/spothero/segment-proxy/bin/segment-proxy /segment-proxy

# Declare the port on which the webserver will be exposed.
# As we're going to run the executable as an unprivileged user, we can't bind
# to ports below 1024.
EXPOSE 8080
EXPOSE 8081

# Perform any further action as an unprivileged user.
USER nobody:nobody

# Run the compiled binary.
CMD ["/segment-proxy"]