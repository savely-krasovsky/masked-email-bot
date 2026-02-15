FROM docker.io/golang:1.26-alpine3.23 AS build

RUN apk add --no-cache build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -v \
    -ldflags="-s -w -extldflags '-static'" \
    -o /app/bot \
    ./cmd/masked-email-bot

FROM docker.io/alpine:3.23

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /app/bot /usr/local/bin/bot

CMD ["/usr/local/bin/bot"]