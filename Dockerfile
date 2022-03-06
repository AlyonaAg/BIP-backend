FROM golang:latest

RUN go version
WORKDIR /back

COPY ./ ./
RUN go mod vendor
RUN go mod download

EXPOSE 8080

RUN go build -o BIP_backend ./cmd/apiserver/main.go
CMD ["./BIP_backend"]