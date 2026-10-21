# Stage 1: Build Frontend UI
FROM node:22-alpine AS ui-builder
WORKDIR /app/ui
COPY ui/package*.json ./
RUN npm ci
COPY ui/ ./
RUN npm run build

# Stage 2: Build Go Backend
FROM golang:1.24-alpine AS server-builder
WORKDIR /app/service
COPY service/go.mod service/go.sum ./
RUN go mod download
COPY service/ ./
# Copy built UI into server embed directory
COPY --from=ui-builder /app/ui/dist /app/service/cmd/server/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /kdb-service ./cmd/server

# Stage 3: Minimal Runtime Container
FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# Persistent database storage directory
RUN mkdir -p /data/database
ENV DATABASE_PATH=/data/database/kdb.db
ENV PORT=8080

COPY --from=server-builder /kdb-service /app/kdb-service

EXPOSE 8080
VOLUME ["/data/database"]

ENTRYPOINT ["/app/kdb-service"]
