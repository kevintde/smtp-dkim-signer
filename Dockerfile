FROM golang:alpine as build
RUN apk --no-cache --update upgrade && apk --no-cache add git

ADD . /go/src/github.com/mback2k/smtp-dkim-signer
WORKDIR /go/src/github.com/mback2k/smtp-dkim-signer

RUN go get
RUN go build -ldflags="-s -w"
RUN chmod +x smtp-dkim-signer

FROM alpine:latest
RUN apk --no-cache --update upgrade && apk --no-cache add ca-certificates

COPY --from=build /go/src/github.com/mback2k/smtp-dkim-signer/smtp-dkim-signer /usr/local/bin/smtp-dkim-signer

RUN addgroup -S serve
RUN adduser -h /data -S -D -G serve serve

WORKDIR /data
USER serve

CMD [ "/usr/local/bin/smtp-dkim-signer" ]
