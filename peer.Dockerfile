FROM golang:1.18

WORKDIR /go/src/app
COPY ./ .

RUN go mod tidy

RUN cd peer && make

RUN cp peer/demedia-peer /usr/local/bin/

ENTRYPOINT ["demedia-peer"]
