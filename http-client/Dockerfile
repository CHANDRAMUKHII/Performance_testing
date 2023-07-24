FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o main

ENV SERVICE_URL=127.0.0.1:57050

CMD ["./main"]