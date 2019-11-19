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
COPY internal internal
COPY pkg pkg
RUN CGO_ENABLED=0 go install github.com/openaustralia/morph-ng/cmd/clay

# Create unprivileged user
RUN adduser -D clay

####
#### clay-run
####
FROM mlandauer/herokuish:for-morph-ng as clay-run

RUN apt-get update && apt-get install -y libsqlite3-dev

# Add prerun script which will disable output buffering for ruby
ADD build/package/prerun.rb /usr/local/lib/prerun.rb

# Add perl buildpack
RUN /bin/herokuish buildpack install https://github.com/miyagawa/heroku-buildpack-perl.git 2da7480a8339f01968ce3979655555a0ade20564

RUN apt-get install -y jq time bc net-tools

ADD build/package/start.sh /bin/
RUN chmod +x /bin/start.sh

COPY --from=builder /go/bin/clay /bin/clay

####
#### clay-server
####
# Now make the runtime image
FROM scratch as clay-server

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/clay /go/bin/clay

# Run unprivileged
USER clay

CMD ["/go/bin/clay", "server"]
