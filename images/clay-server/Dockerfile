# TODO: Use multi-stage docker build for go app
# TODO: Make minimal docker image

FROM golang:1.12.7

WORKDIR /go/src/app
ENV GO111MODULE on

# Downloads the dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Compiles the dependencies of empty.go
COPY empty.go .
RUN go build empty.go
RUN rm empty.go

# Compile the actual thing
COPY main.go .
RUN go install -v

CMD ["app"]
