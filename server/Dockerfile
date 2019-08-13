FROM golang:1.12.7

WORKDIR /go/src/app
ENV GO111MODULE on
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY empty.go .
RUN go run empty.go

COPY main.go .
RUN go install -v

CMD ["app"]
