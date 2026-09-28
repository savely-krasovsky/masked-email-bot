FROM docker.io/golang:1.26-alpine3.23 AS build

ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -v \
    -ldflags="-s -w" \
    -o /app/bot \
    ./cmd/masked-email-bot

FROM docker.io/alpine:3.23

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /app/bot /usr/local/bin/bot

CMD ["/usr/local/bin/bot"]