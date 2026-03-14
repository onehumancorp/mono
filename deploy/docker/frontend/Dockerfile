FROM node:24-bookworm-slim AS frontend_builder

WORKDIR /frontend
COPY srcs/frontend/package.json srcs/frontend/package-lock.json* ./
RUN npm install
COPY srcs/frontend ./
RUN npm run build

FROM golang:1.25-bookworm AS server_builder

WORKDIR /src
COPY go.mod ./
COPY srcs ./srcs
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/frontend-server ./srcs/frontend/server/cmd/frontend

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=server_builder /out/frontend-server /app/frontend-server
COPY --from=frontend_builder /frontend/dist /app/static

ENV FRONTEND_STATIC_DIR=/app/static
ENV FRONTEND_ADDR=:8081
EXPOSE 8081

ENTRYPOINT ["/app/frontend-server"]