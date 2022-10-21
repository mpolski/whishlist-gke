FROM golang:1.14 AS builder
WORKDIR /go/src/app
COPY . .

RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o whishlist *.go

FROM scratch
WORKDIR /usr/src/
COPY --from=builder /go/src/app .
EXPOSE 8080

ENTRYPOINT [ "./whishlist" ]