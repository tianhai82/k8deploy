# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
FROM golang:1.16 as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies using go modules.
# Allows container builds to reuse downloaded dependencies.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Build the binary.
# -mod=readonly ensures immutable go.mod and go.sum in container builds.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -v -o k8deploy

FROM alpine:3.4
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/k8deploy /app/k8deploy

ENTRYPOINT ["./k8deploy"]
