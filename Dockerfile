FROM golang:1.10.3-alpine3.8 AS BUILD

MAINTAINER CMGS <ilskdw@gmail.com>

# make binary
RUN apk add --no-cache git ca-certificates curl make alpine-sdk linux-headers \
    && curl https://glide.sh/get | sh \
    && go get -d github.com/projecteru2/minions
WORKDIR /go/src/github.com/projecteru2/minions
RUN make build && ./eru-minions --version

FROM alpine:3.8

MAINTAINER CMGS <ilskdw@gmail.com>

RUN mkdir /etc/eru/
COPY --from=BUILD /go/src/github.com/projecteru2/minions/eru-minions /usr/bin/eru-minions
