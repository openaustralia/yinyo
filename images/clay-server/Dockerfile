FROM golang:1.12.9-alpine3.10 as builder

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
COPY . .
RUN CGO_ENABLED=0 go install

# Now make the runtime image
FROM scratch

COPY --from=builder /go/bin/clay /go/bin/clay

# TODO: Run this is an unprivileged user
CMD ["/go/bin/clay"]
