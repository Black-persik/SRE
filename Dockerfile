FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sre_tbank ./internal/cmd/sre_tbank

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && adduser -D -H -u 10001 app

WORKDIR /app
COPY --from=build /out/sre_tbank /app/sre_tbank

USER app
EXPOSE 8080

ENTRYPOINT ["/app/sre_tbank"]
