FROM golang:1.22.4-alpine AS builder
WORKDIR /app

COPY Backend/Go-static-server/go.mod Backend/Go-static-server/go.sum Backend/Go-static-server/
RUN cd Backend/Go-static-server && go mod download

COPY Backend/Go-static-server Backend/Go-static-server
COPY ./Frontend ./Frontend
COPY ./config.yaml ./config.yaml

RUN cd Backend/Go-static-server && CGO_ENABLED=0 go build -ldflags='-extldflags "-static"' -o ./compiled ./cmd

FROM alpine
COPY --from=builder /app/Backend/Go-static-server/compiled /app/Backend/Go-static-server/compiled
COPY --from=builder /app/Frontend /app/Frontend
COPY --from=builder /app/config.yaml /app/config.yaml

ENV ROOT_PATH=app
EXPOSE 8081

ENTRYPOINT ["/app/Backend/Go-static-server/compiled"]
