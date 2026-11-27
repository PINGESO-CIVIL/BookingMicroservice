FROM node:22.22.0-alpine3.23@sha256:e4bf2a82ad0a4037d28035ae71529873c069b13eb0455466ae0bc13363826e34 AS fixtures
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY apps/web/package.json apps/web/package.json
RUN corepack enable && pnpm install --frozen-lockfile --ignore-scripts
COPY apps/web/ apps/web/
COPY contracts/openapi.yaml contracts/openapi.yaml
RUN pnpm --dir apps/web run contracts

FROM golang:1.26.5-alpine3.23@sha256:622e56dbc11a8cfe87cafa2331e9a201877271cbff918af53d3be315f3da88cc AS go-deps
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

FROM go-deps AS backend-build
COPY apps/api/cmd/api ./cmd/api
COPY apps/api/internal/ ./internal/
COPY --from=fixtures /app/apps/api/internal/store/fixtures.json ./internal/store/fixtures.json
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/bookingms-api ./cmd/api

FROM swaggerapi/swagger-ui:v5.32.14@sha256:3d93169968848d371a6a56ca1ab18b47a8906ba461b8eba0688866354f5431d5 AS swagger

FROM gcr.io/distroless/static-debian13:nonroot@sha256:f7f8f729987ad0fdf6b05eeeae94b26e6a0f613bdf46feea7fc40f7bd72953e6 AS backend
COPY --from=backend-build /out/bookingms-api /bookingms-api
COPY --from=swagger /usr/share/nginx/html/swagger-ui.css /usr/share/nginx/html/swagger-ui-bundle.js /usr/share/nginx/html/swagger-ui-bundle.js.LICENSE.txt /docs/
COPY contracts/openapi.yaml /docs/openapi.yaml
USER 65532:65532
ENV MOCK_ADDR=0.0.0.0:8080
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=4s --start-period=5s --retries=12 CMD ["/bookingms-api", "healthcheck"]
ENTRYPOINT ["/bookingms-api"]
