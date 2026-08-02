FROM arigaio/atlas:latest AS atlas

FROM alpine:3.21 AS final

RUN apk update && apk upgrade --no-cache && apk add --no-cache bash

# Create non-root user
RUN adduser -D -s /bin/sh appuser

WORKDIR /usr/src/app

COPY --from=atlas /atlas /usr/local/bin/atlas
RUN chmod +rx /usr/local/bin/atlas

COPY migrations /usr/src/app/migrations
COPY ./migrate/run.sh .
RUN chmod +x run.sh

# Change ownership to appuser
RUN chown -R appuser:appuser /usr/src/app

# Switch to non-root user
USER appuser

CMD ["./run.sh"]
