FROM golang:1.16-buster AS builder

WORKDIR /app

COPY go.mod ./
COPY cmd/main.go ./cmd/main.go

RUN go build -o /server ./cmd/main.go

FROM gcr.io/distroless/base-debian10

COPY --from=builder /server /server

EXPOSE 8000

USER nonroot:nonroot

ENTRYPOINT ["/server"]

