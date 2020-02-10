FROM golang:1.13

WORKDIR /
ENV GOPATH /go

COPY . /go/src/reservation-api
RUN go get github.com/go-sql-driver/mysql
RUN go get github.com/go-gorp/gorp

WORKDIR /go/src/reservation-api
RUN go build -o bin/reservation-api
RUN cp bin/reservation-api /usr/local/bin/

CMD ["reservation-api"]