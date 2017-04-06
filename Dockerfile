FROM golang:1.8

EXPOSE 3000

ADD . /go/src/fknsrs.biz/p/don

RUN go install fknsrs.biz/p/don

CMD don
