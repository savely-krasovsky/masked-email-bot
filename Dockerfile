FROM docker.io/golang:1.25-alpine3.22 AS build

WORKDIR /usr/local/src/masked-email-bot

COPY go.mod go.sum ./
RUN go mod download && go mod verify

RUN apk add shadow==4.17.3-r0 gcc==14.2.0-r6 musl-dev==1.2.5-r10 && useradd -u 10001 gopher

COPY . .

RUN go build -v --ldflags '-linkmode external -extldflags=-static' -o /usr/local/bin/masked-email-bot ./cmd/masked-email-bot

FROM docker.io/alpine:3.22

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /usr/local/bin/masked-email-bot /usr/local/bin/masked-email-bot

USER gopher

CMD ["masked-email-bot"]
