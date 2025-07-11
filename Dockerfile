FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/scheduler ./cmd/scheduler

FROM alpine:latest

COPY --from=builder /bin/api /bin/api
COPY --from=builder /bin/worker /bin/worker
COPY --from=builder /bin/scheduler /bin/scheduler

EXPOSE 8080

CMD ["/bin/api"]