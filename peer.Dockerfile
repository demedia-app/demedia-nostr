FROM golang:1.18

WORKDIR /go/src/app
COPY ./ .

RUN go mod tidy

RUN cd peer && make

RUN cp peer/peer-demedia /usr/local/bin/

EXPOSE 8080

ENTRYPOINT ["peer-demedia"]
