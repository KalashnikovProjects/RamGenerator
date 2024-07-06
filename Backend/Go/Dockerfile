FROM golang:1.22.4-alpine AS builder
WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags='-extldflags "-static"' -o compiled ./cmd

FROM scratch
COPY --from=builder /app/compiled /app/compiled

EXPOSE 80
ENTRYPOINT ["/app/compiled"]
