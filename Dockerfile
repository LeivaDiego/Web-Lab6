FROM golang:1.21

WORKDIR /app

COPY go.mod ./
RUN go mod tidy

COPY main.go ./
COPY database ./database

RUN go mod tidy
RUN go build -o server .

EXPOSE 8080

CMD ["./server"]
