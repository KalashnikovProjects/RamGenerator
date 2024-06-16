FROM golang:latest AS builder
WORKDIR /app

COPY ../Backend/Go/ .

RUN go mod download

RUN CGO_ENABLED=0 go build -ldflags='-extldflags "-static"' -o torun ./cmd

# Final stage
FROM scratch
COPY --from=builder ./app/torun ./app/torun

EXPOSE 8080
CMD ["/app/torun"]