FROM golang:1.25-bookworm AS builder

WORKDIR /src
COPY go.mod ./
COPY srcs ./srcs

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/ohc ./srcs/cmd/ohc

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /out/ohc /app/ohc

EXPOSE 8080
ENTRYPOINT ["/app/ohc"]