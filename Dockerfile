FROM docker.io/golang:1.20-alpine3.18 AS build

WORKDIR /usr/local/src/masked-email-bot

COPY go.mod go.sum ./
RUN go mod download && go mod verify

RUN apk add shadow==4.13-r2 gcc==12.2.1_git20220924-r10 musl-dev==1.2.4-r0 && useradd -u 10001 gopher

COPY . .

RUN go build -v --ldflags '-linkmode external -extldflags=-static' -o /usr/local/bin/masked-email-bot ./cmd/masked-email-bot

FROM docker.io/alpine:3.18

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /usr/local/bin/masked-email-bot /usr/local/bin/masked-email-bot

USER gopher

CMD ["masked-email-bot"]
