FROM golang:1.27.3-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/app /usr/local/bin/app

USER nobody

ENTRYPOINT ["/usr/local/bin/app"]
