FROM node:20-bookworm-slim AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci --prefer-offline --no-audit --no-fund
COPY web/ ./
RUN npm run build

FROM golang:1.20-bookworm AS go
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags '-s -w' -o /out/store ./cmd/store

FROM gcr.io/distroless/static-debian12
COPY --from=go /out/store /usr/local/bin/store
ENTRYPOINT ["/usr/local/bin/store"]
CMD ["start", "/var/www/store/api.socket", "/var/www/store/secret.yaml"]
