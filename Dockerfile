FROM golang:1.21

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /where-is-the-painting

EXPOSE 8080

CMD [ "/where-is-the-painting" ]

