FROM golang:1.18

WORKDIR /go/src/gia-app/common

COPY . .

RUN go mod tidy

ENTRYPOINT ["go", "test", "-v", "./..."]