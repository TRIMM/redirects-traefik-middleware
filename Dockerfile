ARG GO_VERSION

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh gcc libc-dev

ENV GOPATH /app

RUN mkdir -p /app/redirects-traefik-middleware
ADD ./go.* /app/redirects-traefik-middleware/
#ADD redirects.db /app/redirects-traefik-middleware/
ADD ./cmd  /app/redirects-traefik-middleware/cmd
ADD ./api  /app/redirects-traefik-middleware/api
ADD ./internal /app/redirects-traefik-middleware/internal

WORKDIR /app/redirects-traefik-middleware
RUN go mod download

RUN CGO_ENABLED=1 go build -o app ./cmd

FROM alpine:latest

COPY --from=builder /app/redirects-traefik-middleware/app .

EXPOSE 8080

CMD ["./app"]
