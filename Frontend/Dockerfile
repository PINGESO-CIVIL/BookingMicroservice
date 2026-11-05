FROM node:24.22.2-alpine3.23 AS build
WORKDIR /app/Frontend
COPY Frontend/package.json Frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY Frontend/ ./
COPY contracts/openapi/ /app/contracts/openapi/
# Regenerate types/examples from the contract inside the build, never on the host.
RUN npm run contracts && npm run build

FROM nginx:1.30.6-alpine
# Refresh the runtime OS packages; --pull/--no-cache rebuilds refresh security fixes.
RUN apk upgrade --no-cache
COPY Frontend/nginx.conf /etc/nginx/nginx.conf
COPY --from=build /app/Frontend/dist /usr/share/nginx/html
USER 101:101
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=12 CMD ["wget", "--spider", "-q", "http://127.0.0.1:8080/healthz"]
ENTRYPOINT ["nginx"]
CMD ["-g", "daemon off;"]
