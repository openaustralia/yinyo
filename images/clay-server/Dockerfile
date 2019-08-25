# TODO: Use multi-stage docker build for go app
# TODO: Make minimal docker image

FROM golang:1.12.9-alpine3.10

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

WORKDIR /go/src/app
ENV GO111MODULE on

# Downloads the dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Compiles the dependencies of empty.go
COPY empty/empty.go .
RUN go build empty.go
RUN rm empty.go

# Compile the actual thing
COPY *.go ./
RUN go install

CMD ["app"]
