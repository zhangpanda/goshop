# ---- Go backend ----
FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o bin/goshop ./cmd/server/

# ---- Final image (backend only) ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=backend /app/bin/goshop .
COPY --from=backend /app/config.yaml.example ./config.yaml
COPY --from=backend /app/uploads/seed ./uploads/seed
COPY --from=backend /app/static ./static

EXPOSE 8080
CMD ["./goshop"]
