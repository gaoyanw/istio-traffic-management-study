FROM golang:1.16-buster AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY pkg/ ./pkg/

COPY cmd/bookstoreserver/main.go ./cmd/bookstoreserver/main.go

RUN go build -o /server ./cmd/bookstoreserver/main.go

FROM gcr.io/distroless/base-debian10

COPY --from=builder /server /server

EXPOSE 3443

USER nonroot:nonroot

ENTRYPOINT ["/server"]

