FROM golang:1.19 as builder

WORKDIR /go/src/app
COPY ./ .

RUN go mod tidy

RUN cd peer && make

FROM alpine:3.14

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/app/peer/demedia-peer /usr/local/bin/

ENTRYPOINT ["demedia-peer"]
