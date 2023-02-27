FROM golang:1.16-buster AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

COPY cmd/extauthzserver/main.go ./cmd/extauthzserver/main.go

RUN go build -o /server ./cmd/extauthzserver/main.go

FROM gcr.io/distroless/base-debian10

COPY --from=builder /server /server

EXPOSE 9000
EXPOSE 8000


USER nonroot:nonroot

ENTRYPOINT ["/server"]