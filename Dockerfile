FROM golang:1.9-alpine

ENV APP_PATH=github.com/mat285/pinger
ENV APP_ROOT=/go/src/${APP_PATH}

ADD vendor ${APP_ROOT}/vendor
ADD main.go ${APP_ROOT}/main.go

RUN go install ${APP_PATH}

ENTRYPOINT ["/go/bin/pinger"]
