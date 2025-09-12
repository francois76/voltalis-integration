FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY . .
RUN go build -o /voltalis ./cmd/voltalis

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /voltalis /app/voltalis
COPY run.sh /app/run.sh
RUN chmod +x /app/run.sh
CMD ["/app/run.sh"]
