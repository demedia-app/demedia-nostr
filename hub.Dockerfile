FROM golang:1.19 as builder

WORKDIR /go/src/app
COPY ./ .

RUN go mod tidy

RUN cd hub && make

FROM alpine:3.14

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/app/hub/demedia-hub /usr/local/bin/

ENTRYPOINT ["demedia-hub"]
