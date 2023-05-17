FROM golang:1.18

WORKDIR /go/src/app
COPY ./ .

RUN go mod tidy

RUN cd hub && make

RUN cp hub/hub-demedia /usr/local/bin/

EXPOSE 8080

ENTRYPOINT ["hub-demedia"]
