FROM golang:1.24-alpine as builder

RUN apk update && apk add --no-cache ca-certificates git

WORKDIR /app

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/scheduler ./cmd/scheduler
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/archive-monitor ./cmd/archive-monitor

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl procps

WORKDIR /app

RUN addgroup -S nonroot && adduser -S nonroot -G nonroot

COPY --from=builder /go/bin/migrate /app/migrate
COPY --from=builder /app/api /app/api
COPY --from=builder /app/worker /app/worker
COPY --from=builder /app/scheduler /app/scheduler
COPY --from=builder /app/archive-monitor /app/archive-monitor
COPY --from=builder /app/migrations /app/migrations

RUN chown -R nonroot:nonroot /app

EXPOSE 8080
USER nonroot

CMD ["/app/api"]