FROM golang:latest

RUN go version
WORKDIR /back

COPY ./ ./
RUN go install github.com/githubnemo/CompileDaemon@latest
RUN go mod tidy
RUN go mod download

EXPOSE 8080

ENTRYPOINT CompileDaemon --build="go build -o BIP_backend ./cmd/apiserver/main.go" --command=./BIP_backend
#RUN go build -o BIP_backend ./cmd/apiserver/main.go
#CMD ["./BIP_backend"]