FROM golang:1.11.2 AS builder

COPY go.mod go.sum /src/
WORKDIR /src

RUN GO111MODULE=on go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app server/main.go

# Runtime image
FROM alpine:3.8 as base
COPY --from=builder /src/app /bin/app
ENTRYPOINT ["/bin/app"]