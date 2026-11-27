FROM node:22.22.0-alpine3.23@sha256:e4bf2a82ad0a4037d28035ae71529873c069b13eb0455466ae0bc13363826e34 AS build
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY apps/web/package.json apps/web/package.json
RUN corepack enable && pnpm install --frozen-lockfile --ignore-scripts
COPY apps/web/ apps/web/
COPY contracts/openapi.yaml contracts/openapi.yaml
RUN pnpm --dir apps/web run contracts && pnpm --dir apps/web build

FROM golang:1.26.5-alpine3.23@sha256:622e56dbc11a8cfe87cafa2331e9a201877271cbff918af53d3be315f3da88cc AS server
COPY apps/web/server.go /src/server.go
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/web /src/server.go

FROM gcr.io/distroless/static-debian13:nonroot@sha256:f7f8f729987ad0fdf6b05eeeae94b26e6a0f613bdf46feea7fc40f7bd72953e6
COPY --from=server /out/web /web
COPY --from=build /app/apps/web/dist /app/dist
USER 65532:65532
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=12 CMD ["/web", "healthcheck"]
ENTRYPOINT ["/web"]
