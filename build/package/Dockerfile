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
RUN CGO_ENABLED=0 go install github.com/openaustralia/yinyo/cmd/yinyo

# Create unprivileged user
RUN adduser -D yinyo

####
#### yinyo-run
####
FROM mlandauer/herokuish:for-morph-ng as yinyo-run

RUN apt-get update && apt-get install -y libsqlite3-dev

# Add prerun script which will disable output buffering for ruby
ADD build/package/prerun.rb /usr/local/lib/prerun.rb

# Add perl buildpack
RUN /bin/herokuish buildpack install https://github.com/miyagawa/heroku-buildpack-perl.git 2da7480a8339f01968ce3979655555a0ade20564

ADD build/package/start.sh /bin/
RUN chmod +x /bin/start.sh

COPY --from=builder /go/bin/yinyo /bin/yinyo

# Install geckdriver for firefox
RUN wget https://github.com/mozilla/geckodriver/releases/download/v0.26.0/geckodriver-v0.26.0-linux64.tar.gz && \
    tar xf geckodriver-v0.26.0-linux64.tar.gz && \
    rm geckodriver-v0.26.0-linux64.tar.gz && \
    mv geckodriver /usr/local/bin

# Install firefox
# TODO: Install a specific version of firefox so that we can actively upgrade firefox
RUN apt-get -y install firefox

####
#### yinyo-server
####
# Now make the runtime image
FROM scratch as yinyo-server

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/yinyo /go/bin/yinyo
COPY --from=builder --chown=yinyo:0 /tmp /tmp

# Run unprivileged
USER yinyo

CMD ["/go/bin/yinyo", "server"]
