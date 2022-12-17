FROM gcr.io/distroless/base

COPY ./templates /templates
COPY ./ent-web /ent-web

ENV PORT=8080
ENV GIN_MODE=release
ENV ENABLE_MEMCACHE=1
ENV ENABLE_BIGQUERY=1

ENTRYPOINT [ "/ent-web" ]
