FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /leaseweb-exporter ./cmd/leaseweb_exporter

# --- Stage 2: Final ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -u 1000 exporter
USER exporter

WORKDIR /home/exporter
COPY --from=builder /leaseweb-exporter .

EXPOSE 9112

ENTRYPOINT ["./leaseweb-exporter"]
