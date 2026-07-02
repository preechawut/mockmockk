# syntax=docker/dockerfile:1

# ---------- build ----------
FROM golang:1.25-alpine AS build

WORKDIR /src

# cache deps
COPY go.mod go.sum ./
RUN go mod download

# build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/api

# ---------- runtime ----------
FROM alpine:3.20

WORKDIR /app

# migrations are read at startup from ./db/migrations
COPY --from=build /out/server /app/server
COPY --from=build /src/db/migrations /app/db/migrations

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/server"]
