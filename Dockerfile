FROM golang:1.7-alpine

ENV PATH $PATH:/opt/flamegraph

RUN apk --update add git && \
    apk add curl && \
    curl -OL https://github.com/Masterminds/glide/releases/download/v0.12.3/glide-v0.12.3-linux-amd64.tar.gz && \
    tar -xzf glide-v0.12.3-linux-amd64.tar.gz && \
    mv linux-amd64/glide /usr/bin && \
    apk add perl && \
    git clone --depth=1 https://github.com/brendangregg/FlameGraph.git /opt/flamegraph

COPY . /go/src/github.com/uber/go-torch

RUN cd /go/src/github.com/uber/go-torch && glide install && go install ./...

ENTRYPOINT ["go-torch"]
