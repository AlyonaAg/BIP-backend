FROM golang:latest

RUN go version
WORKDIR /back

COPY go.mod ./
RUN go mod download
COPY ./ ./

EXPOSE 8080

RUN go build -o BIP_backend ./cmd/apiserver/main.go
CMD ["./BIP_backend"]
