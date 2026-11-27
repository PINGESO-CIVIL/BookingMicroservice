FROM node:24.22.2-alpine3.23 AS fixtures
WORKDIR /app/apps/web
COPY apps/web/package.json apps/web/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY apps/web/ ./
COPY contracts/openapi.yaml /app/contracts/openapi.yaml
RUN npm run contracts

FROM golang:1.27.3-alpine3.23 AS go-deps
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

FROM go-deps AS backend-build
COPY apps/api/cmd/api ./cmd/api
COPY apps/api/internal/ ./internal/
COPY --from=fixtures /app/apps/api/internal/store/fixtures.json ./internal/store/fixtures.json
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/bookingms-api ./cmd/api

FROM swaggerapi/swagger-ui:v5.32.14@sha256:3d93169968848d371a6a56ca1ab18b47a8906ba461b8eba0688866354f5431d5 AS swagger

FROM scratch AS backend
COPY --from=backend-build /out/bookingms-api /bookingms-api
COPY --from=swagger /usr/share/nginx/html/swagger-ui.css /usr/share/nginx/html/swagger-ui-bundle.js /usr/share/nginx/html/swagger-ui-bundle.js.LICENSE.txt /docs/
COPY contracts/openapi.yaml /docs/openapi.yaml
USER 65532:65532
ENV MOCK_ADDR=0.0.0.0:8080
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=4s --start-period=5s --retries=12 CMD ["/bookingms-api", "healthcheck"]
ENTRYPOINT ["/bookingms-api"]
