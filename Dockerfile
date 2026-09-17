FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/copilot-proxy-go .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -u 10001 app

COPY --from=builder /out/copilot-proxy-go /usr/local/bin/copilot-proxy-go

USER app

EXPOSE 4141

ENTRYPOINT ["copilot-proxy-go"]
CMD ["start"]