FROM golang:1.25-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
COPY configs ./configs
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/main ./cmd/server

FROM node:20-slim AS frontend-builder
WORKDIR /frontend
COPY frontend/package*.json ./
COPY frontend/.npmrc ./
RUN npm ci --no-audit --prefer-offline
COPY frontend/svelte.config.js ./
COPY frontend/vite.config.ts ./
COPY frontend/tsconfig.json ./
COPY frontend/src ./src
COPY frontend/static ./static
RUN npm run build

FROM alpine:latest
RUN apk add --no-cache tzdata ca-certificates && \
    mkdir -p /var/log/teweicun
WORKDIR /app
COPY --from=backend-builder /out/main /app/main
COPY --from=backend-builder /app/configs /app/configs
COPY --from=frontend-builder /frontend/build /app/frontend-dist
EXPOSE 8080
CMD ["./main", "-c", "/app/configs/config.yaml"]
