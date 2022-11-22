FROM alpine

# create and use non-root user
RUN adduser \
  --disabled-password \
  --home /app \
  --gecos '' app \
  && chown -R app /app

USER app

WORKDIR /app
COPY ./.output/app ./

ENTRYPOINT [ "./confluent-gateway" ]