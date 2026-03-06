# --- Stage 1: Frontend build ---
FROM node:22-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npx svelte-kit sync && npm run build

# --- Stage 2: Backend build ---
FROM golang:1.23-alpine AS backend
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/build ./web/dist
ARG BUILD_VERSION=dev
ARG BUILD_COMMIT=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${BUILD_VERSION} -X main.commit=${BUILD_COMMIT}" \
    -o /memento ./cmd/memento

# --- Stage 3: Runtime ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S memento && adduser -S -G memento memento \
    && mkdir -p /data/attachments && chown -R memento:memento /data
COPY --from=backend /memento /usr/local/bin/memento
USER memento
EXPOSE 3000
ENTRYPOINT ["memento"]
