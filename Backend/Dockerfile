FROM node:24.22.2-alpine3.23 AS fixtures
WORKDIR /app/Frontend
COPY Frontend/package.json Frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY Frontend/ ./
COPY contracts/openapi/ /app/contracts/openapi/
RUN npm run contracts

FROM golang:1.27.3-alpine3.23 AS build
WORKDIR /src
COPY Backend/go.mod ./
COPY Backend/cmd/mock-api ./cmd/mock-api
COPY Backend/internal/mockapi ./internal/mockapi
COPY --from=fixtures /app/Backend/internal/mockapi/fixtures.json ./internal/mockapi/fixtures.json
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/mock-api ./cmd/mock-api

FROM scratch
COPY --from=build /out/mock-api /mock-api
USER 65532:65532
ENV MOCK_ADDR=0.0.0.0:4010
EXPOSE 4010
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=12 CMD ["/mock-api", "healthcheck"]
ENTRYPOINT ["/mock-api"]
