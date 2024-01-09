FROM golang:1.21.4 AS builder

WORKDIR /go/src/app

COPY . .

# RUN go env -w GOPROXY=https://goproxy.cn,direct

RUN go get .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o service .

FROM alpine:3.19.0

COPY --from=builder /go/src/app/service /service

CMD ["/service"]