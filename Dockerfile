FROM golang:1.24

WORKDIR /go/src/app

COPY . .

EXPOSE 8080

RUN go build -o main scrapper.go

CMD ["./main"]