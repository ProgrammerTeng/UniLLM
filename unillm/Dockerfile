# ---- Stage 1: Build Go backend ----
FROM golang:1.26-alpine AS go-builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /unillm ./cmd/server


# ---- Stage 2: Build Next.js frontend ----
FROM node:22-alpine AS web-builder

WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm ci --prefer-offline
COPY web/ .
RUN npm run build


# ---- Stage 3: Production Go backend ----
FROM alpine:3.21 AS backend

RUN apk --no-cache add ca-certificates tzdata
COPY --from=go-builder /unillm /unillm

EXPOSE 8080
ENTRYPOINT ["/unillm"]


# ---- Stage 4: Production Next.js frontend ----
FROM node:22-alpine AS frontend

WORKDIR /app

# Copy standalone output
COPY --from=web-builder /app/web/.next/standalone ./
COPY --from=web-builder /app/web/.next/static ./.next/static
COPY --from=web-builder /app/web/public ./public

ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

EXPOSE 3000
CMD ["node", "server.js"]
