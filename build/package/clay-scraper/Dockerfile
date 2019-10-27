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

# Compile the actual thing
# Only copying across the bits we actually need so that skaffold isn't overeager on
# rebuilding everything when another (not relevant) file changes
COPY cmd cmd
COPY internal internal
COPY pkg pkg
RUN CGO_ENABLED=0 go install github.com/openaustralia/morph-ng/cmd/clay-run

FROM mlandauer/herokuish:for-morph-ng

RUN apt-get update && apt-get install -y libsqlite3-dev

# Add prerun script which will disable output buffering for ruby
ADD build/package/clay-scraper/prerun.rb /usr/local/lib/prerun.rb

# Add perl buildpack
RUN /bin/herokuish buildpack install https://github.com/miyagawa/heroku-buildpack-perl.git 2da7480a8339f01968ce3979655555a0ade20564

RUN apt-get install -y jq time bc net-tools

ADD build/package/clay-scraper/run.sh build/package/clay-scraper/start.sh /bin/
RUN chmod +x /bin/run.sh /bin/start.sh

COPY --from=builder /go/bin/clay-run /bin/clay-run
