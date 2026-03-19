FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY frontend-server /app/frontend_server
COPY static /app/static
ENV FRONTEND_STATIC_DIR=/app/static
ENV FRONTEND_ADDR=:8081
EXPOSE 8081
ENTRYPOINT ["/app/frontend_server"]
