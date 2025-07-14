FROM golang:1.24-alpine as builder

RUN apk update && apk add --no-cache ca-certificates git && \
    apk upgrade && \
    rm -rf /var/cache/apk/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /bin/worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /bin/scheduler ./cmd/scheduler

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/api /bin/api
COPY --from=builder /bin/worker /bin/worker  
COPY --from=builder /bin/scheduler /bin/scheduler

EXPOSE 8080
USER nonroot

CMD ["/bin/api"]