FROM golang:1.26.5-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/exe ./cmd/chat

FROM alpine:3.24
# в alpine нет базы часовых поясов, без неё time.LoadLocation упадёт на всём кроме UTC
RUN apk add --no-cache tzdata
WORKDIR /app
COPY --from=builder /app/exe /app/exe
CMD ["/app/exe"]
