FROM golang:1.22.4-alpine AS builder
WORKDIR /app

COPY cdn cdn

RUN cd cdn && go mod download && CGO_ENABLED=0 go build -ldflags='-extldflags "-static"' -o compiled .


FROM alpine

COPY --from=builder /app/cdn/compiled /app/cdn/compiled
ENV ROOT_PATH=app

EXPOSE 8080

ENTRYPOINT ["/app/cdn/compiled"]
