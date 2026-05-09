FROM golang:1.25-alpine

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o cert_exporter

FROM alpine:3.18

COPY --from=0 /app/cert_exporter /app/cert_exporter

EXPOSE 9091

CMD ["/app/cert_exporter"]
