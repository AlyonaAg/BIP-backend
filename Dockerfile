FROM golang:latest

RUN go version
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY ./ ./

RUN go build -o BIP_backend ./cmd/apiserver/main.go

CMD ["./BIP_backend"]
