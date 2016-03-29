FROM golang:1.6-alpine

ENV PATH $PATH:/opt/flamegraph

RUN apk --update add git && \
    curl -OL https://github.com/Masterminds/glide/releases/download/0.10.1/glide-0.10.1-linux-amd64.tar.gz && \
    tar -xzf glide-0.10.1-linux-amd64.tar.gz && \
    mv linux-amd64/glide /usr/bin && \
    git clone --depth=1 https://github.com/brendangregg/FlameGraph.git /opt/flamegraph

COPY . /go/src/github.com/uber/go-torch

RUN cd /go/src/github.com/uber/go-torch && glide install && go install ./...

ENTRYPOINT ["go-torch"]
