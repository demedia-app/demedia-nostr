FROM golang:1.18 as builder

WORKDIR /go/src/app
COPY ./ .

RUN go mod tidy

RUN cd hub && make

FROM ubuntu:22.04

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/app/hub/demedia-hub /usr/local/bin/

ENTRYPOINT ["demedia-hub"]
