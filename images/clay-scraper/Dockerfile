# PART ONE
FROM golang:1.12.9 as builder

WORKDIR /go/src/docker-stats-on-exit-shim
RUN git clone https://github.com/delcypher/docker-stats-on-exit-shim.git /go/src/docker-stats-on-exit-shim
RUN git checkout 650e6e9c
RUN git submodule init && git submodule update
RUN go get .
RUN go build

# PART TWO
FROM mlandauer/herokuish:for-morph-ng
MAINTAINER Matthew Landauer <matthew@oaf.org.au>

RUN apt-get update && apt-get install -y libsqlite3-dev

# Add prerun script which will disable output buffering for ruby
ADD prerun.rb /usr/local/lib/prerun.rb

# Add standard Procfiles
ADD Procfile /usr/local/lib

# Add perl buildpack
RUN /bin/herokuish buildpack install https://github.com/miyagawa/heroku-buildpack-perl.git 2da7480a8339f01968ce3979655555a0ade20564

RUN apt-get install -y jq

ADD run.sh start.sh clay.sh /bin/
RUN chmod +x /bin/run.sh /bin/start.sh /bin/clay.sh

COPY --from=builder /go/bin/docker-stats-on-exit-shim /bin/docker-stats-on-exit-shim
