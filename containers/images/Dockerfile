FROM golang:1.25-alpine3.20 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/app ./cmd


FROM alpine:3.20

WORKDIR /src

COPY --from=builder /bin/app /bin/app
COPY --from=builder /src/views /src/views
COPY --from=builder /src/assets /src/assets

ENTRYPOINT ["/bin/app"]
