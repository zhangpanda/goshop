# ---- Go backend ----
FROM golang:1.23-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o bin/goshop ./cmd/server/

# ---- Web frontend ----
FROM node:20-alpine AS web
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

# ---- Admin frontend ----
FROM node:20-alpine AS admin
WORKDIR /app/admin
COPY admin/package.json admin/package-lock.json ./
RUN npm ci
COPY admin/ .
RUN npm run build

# ---- Final image ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=backend /app/bin/goshop .
COPY --from=backend /app/config.yaml.example ./config.yaml
COPY --from=backend /app/uploads/seed ./uploads/seed
COPY --from=web /app/web/.next/standalone ./web/
COPY --from=web /app/web/.next/static ./web/.next/static
COPY --from=web /app/web/public ./web/public
COPY --from=admin /app/admin/.next/standalone ./admin/
COPY --from=admin /app/admin/.next/static ./admin/.next/static

EXPOSE 8080 3000 3001
CMD ["./goshop"]
