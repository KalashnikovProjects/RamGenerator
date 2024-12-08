FROM golang:1.22.4-alpine AS builder
WORKDIR /app

COPY Backend/Go-api/go.mod Backend/Go-static-server/go.sum Backend/Go-api/
RUN cd Backend/Go-api && go mod download

COPY Backend/Go-api Backend/Go-api
COPY ./config.yaml ./config.yaml

RUN cd Backend/Go-api && CGO_ENABLED=0 go build -ldflags='-extldflags "-static"' -o compiled ./cmd

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/Backend/Go-api/compiled /app/Backend/Go-api/compiled
COPY --from=builder /app/config.yaml /app/config.yaml

ENV ROOT_PATH=app
EXPOSE 8082

ENTRYPOINT ["/app/Backend/Go-api/compiled"]
