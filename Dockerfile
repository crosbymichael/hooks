FROM golang:1.3

COPY . /go/src/github.com/crosbymichael/hooks
WORKDIR /go/src/github.com/crosbymichael/hooks
RUN go get -d ./... && go install ./...
ENTRYPOINT ["hooks"]
