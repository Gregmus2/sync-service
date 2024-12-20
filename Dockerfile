FROM gregmus2/golang-base:1.23 as builder

COPY . .

RUN apk add gcc musl-dev

RUN CGO_ENABLED=1 go build -o /go/bin/app ./cmd/app/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /go/bin/app /bin/app

EXPOSE 9000

ENTRYPOINT ["/bin/app"]
