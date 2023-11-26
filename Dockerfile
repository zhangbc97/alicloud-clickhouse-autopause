FROM golang:1.21.4 AS builder

WORKDIR /go/src/app

COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct && go get .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o service .

FROM scratch

COPY --from=builder /go/src/app/service /service

CMD ["/service"]