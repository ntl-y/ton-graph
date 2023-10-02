FROM golang:1.21

WORKDIR /app

ENV GOPATH=/

COPY . .

RUN go mod download
RUN go build -o ton-graph ./main.go

CMD [ "./ton-graph" ]

