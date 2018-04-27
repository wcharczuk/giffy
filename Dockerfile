FROM golang:1-alpine

ENV APP_PATH=github.com/wcharczuk/giffy
ENV APP_ROOT=/go/src/${APP_PATH}

ADD vendor ${APP_ROOT}/vendor 
ADD database ${APP_ROOT}/database
ADD server ${APP_ROOT}/server
ADD _client/dist ${APP_ROOT}/_client/dist
ADD main.go ${APP_ROOT}/main.go

ARG CURRENT_REF
ENV CURRENT_REF ${CURRENT_REF}

RUN go install ${APP_PATH}
WORKDIR ${APP_ROOT}
ENTRYPOINT ["/go/bin/giffy"]
EXPOSE 8080
