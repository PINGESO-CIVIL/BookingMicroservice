# Keep the reviewed UI release and rebuild its OS layer with security updates.
FROM swaggerapi/swagger-ui:v5.32.14@sha256:3d93169968848d371a6a56ca1ab18b47a8906ba461b8eba0688866354f5431d5
USER root
RUN apk upgrade --no-cache
COPY infra/swagger/nginx.conf /etc/nginx/nginx.conf
# The upstream entrypoint writes the generated UI configuration and gzip assets.
RUN chown -R 101:101 /usr/share/nginx/html /etc/nginx/conf.d /etc/nginx/templates
USER 101:101
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=12 CMD ["wget", "--spider", "-q", "http://127.0.0.1:8080/docs/"]
