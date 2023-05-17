FROM golang:1.18

WORKDIR /go/src/app
COPY ./ .

RUN go mod tidy

RUN cd hub && make

RUN cp hub/demedia-hub /usr/local/bin/

ENTRYPOINT ["demedia-hub"]
