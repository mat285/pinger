FROM golang:1.9-alpine

ENV APP_PATH=github.com/mat285/pinger
ENV APP_ROOT=/go/src/${APP_PATH}

RUN go install ${APP_PATH}/pinger

ENTRYPOINT ["/go/bin/pinger"]
