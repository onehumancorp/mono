# User Guide: Deploy Locally with Docker

## Purpose

Start the full app stack locally using Docker images built from Bazel targets.

## Step 1: Build and load images

```bash
bazelisk run //deploy:backend_image_tarball
bazelisk run //deploy:frontend_image_tarball
```

## Step 2: Start compose stack

```bash
docker compose -f deploy/docker-compose.yml up -d
```

## Step 3: Access services

- Backend: `http://localhost:8080`
- Frontend: `http://localhost:8081`

## Step 4: Stop stack

```bash
docker compose -f deploy/docker-compose.yml down
```

## Troubleshooting

- If image not found, rerun Bazel `run` commands to load images.
- If port conflict happens, stop existing local services bound to `8080` or `8081`.
