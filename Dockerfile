# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go build -o /out/stub

FROM alpine:3.20
RUN adduser -D -u 10001 app
USER app
COPY --from=build /out/stub /usr/local/bin/stub
ENTRYPOINT ["/usr/local/bin/stub"]
