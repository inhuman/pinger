FROM golang:1.16-alpine AS builder

ENV CGO_ENABLED=0

ENV TZ=Europe/Moscow

RUN apk --no-cache add ca-certificates tzdata && \
    cp -r -f /usr/share/zoneinfo/$TZ /etc/localtime

WORKDIR /app

COPY . .

RUN go build -mod=vendor -o /pinger ./cmd/pinger

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=builder /etc/localtime /etc/localtime

COPY --from=builder /pinger /pinger

ENTRYPOINT ["/pinger"]

EXPOSE 9000
