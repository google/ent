FROM gcr.io/distroless/base

COPY ./templates /templates
COPY ./ent /ent

ENV PORT=8080
ENV GIN_MODE=release
ENV ENABLE_MEMCACHE=1
ENV ENABLE_BIGQUERY=1

ENTRYPOINT [ "/ent" ]
