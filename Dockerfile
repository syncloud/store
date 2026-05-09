FROM node:20-bookworm-slim AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci --prefer-offline --no-audit --no-fund
COPY web/ ./
RUN npm run build

FROM golang:1.20-bookworm AS go
ARG GIT_SHA=unknown
ARG BUILD_NUMBER=0
ARG BUILD_TIME=unknown
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w \
        -X github.com/syncloud/store/internal/version.GitSha=${GIT_SHA} \
        -X github.com/syncloud/store/internal/version.BuildNumber=${BUILD_NUMBER} \
        -X github.com/syncloud/store/internal/version.BuildTime=${BUILD_TIME}" \
    -o /out/store ./cmd/store
RUN printf 'gitSha=%s\nbuildNumber=%s\nbuildTime=%s\n' "${GIT_SHA}" "${BUILD_NUMBER}" "${BUILD_TIME}" > /out/VERSION

FROM gcr.io/distroless/static-debian12
COPY --from=go /out/store /usr/local/bin/store
COPY --from=go /out/VERSION /VERSION
ENTRYPOINT ["/usr/local/bin/store"]
CMD ["start", "/var/www/store/api.socket", "/var/www/store/secret.yaml"]
