FROM golang:alpine as build

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh gcc musl-dev
ENV GOROOT=/usr/local/go
RUN go get golang.org/x/net/websocket
RUN go get github.com/btcsuite/btcd
RUN go get github.com/btcsuite/btcutil
RUN go get github.com/fatih/color
RUN go get github.com/gertjaap/lit/wire
COPY . /usr/local/go/src/github.com/gertjaap/lit-docker-tester
WORKDIR /usr/local/go/src/github.com/gertjaap/lit-docker-tester
RUN go build

FROM alpine
RUN apk add --no-cache ca-certificates
WORKDIR /app
RUN cd /app
COPY --from=build /usr/local/go/src/github.com/gertjaap/lit-docker-tester/lit-docker-tester /app/bin/lit-docker-tester

EXPOSE 8001

CMD ["bin/lit-docker-tester"]