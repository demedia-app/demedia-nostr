FROM golang:1.18 as builder

WORKDIR /go/src/app
COPY ./ .

RUN go mod tidy

RUN cd hub && make

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

COPY --from=builder /go/src/app/hub/demedia-hub /usr/local/bin/

ENTRYPOINT ["demedia-hub"]
