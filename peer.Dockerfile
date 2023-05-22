FROM golang:1.18 as builder

WORKDIR /go/src/app
COPY ./ .

RUN go mod tidy

RUN cd peer && make

FROM ubuntu:22.04

COPY --from=builder /go/src/app/peer/demedia-peer /usr/local/bin/

ENTRYPOINT ["demedia-peer"]
