FROM golang:1.13.1-alpine3.10 as builder

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
COPY pkg/empty/empty.go .
RUN go build empty.go
RUN rm empty.go

# Compile the actual thing
# Only copying across the bits we actually need so that skaffold isn't overeager on
# rebuilding everything when another (not relevant) file changes
COPY cmd cmd
COPY internal internal
COPY pkg pkg
RUN CGO_ENABLED=0 go install github.com/openaustralia/morph-ng/cmd/clay-server

# Now make the runtime image
FROM scratch

COPY --from=builder /go/bin/clay-server /go/bin/clay-server

# TODO: Run this is an unprivileged user
CMD ["/go/bin/clay-server"]
