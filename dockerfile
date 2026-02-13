FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o apex-motors .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/apex-motors .
COPY static/ ./static/

EXPOSE 5001
CMD ["./apex-motors"]