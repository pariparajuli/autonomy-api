FROM golang:1.13-alpine as build

WORKDIR $GOPATH/github.com/bitmark-inc/autonomy-api

ADD go.mod .

RUN go mod download

ADD . .

RUN go install github.com/bitmark-inc/autonomy-api
RUN go install github.com/bitmark-inc/autonomy-api/schema/command/migrate

# ---

FROM alpine:3.10.3
ARG dist=0.0
COPY --from=build /go/bin/autonomy-api /
COPY --from=build /go/bin/migrate /

ENV AUTONOMY_LOG_LEVEL=INFO
ENV AUTONOMY_SERVER_VERSION=$dist

CMD ["/autonomy-api"]
