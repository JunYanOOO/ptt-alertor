FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ptt-alertor .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=builder /out/ptt-alertor ./ptt-alertor
COPY public/ ./public/

RUN mkdir -p /app/storage/articles /app/storage/users \
    && chown -R app:app /app

USER app

EXPOSE 9090 6060

ENTRYPOINT ["./ptt-alertor"]
