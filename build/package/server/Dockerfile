####
#### builder
####
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
COPY pkg pkg
COPY .git .

RUN GIT_COMMIT=$(git rev-list -1 HEAD) && \
    CGO_ENABLED=0 go install -ldflags "-X main.GitCommit=$GIT_COMMIT" github.com/openaustralia/yinyo/cmd/server

# Create unprivileged user
RUN adduser -D yinyo

####
#### yinyo-server
####
# Now make the runtime image
FROM scratch as yinyo-server

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/server /go/bin/server
COPY --from=builder --chown=yinyo:0 /tmp /tmp

# Run unprivileged
USER yinyo

CMD ["/go/bin/server"]
