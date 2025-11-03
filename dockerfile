FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o publisher .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/publisher .
COPY --from=builder /app/credentials.json .
COPY --from=builder /app/.env .
CMD ["./publisher"]