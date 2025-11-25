FROM golang:1.24 AS builder

WORKDIR /app
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM gcr.io/distroless/base-debian12

COPY --from=builder /app/server .

EXPOSE 3000

ENTRYPOINT ["./server"]
