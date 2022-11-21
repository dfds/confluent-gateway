FROM alpine

WORKDIR /app
COPY ./.output/app ./

# create and use non-root user
RUN adduser \
  --disabled-password \
  --home /app \
  --gecos '' app \
  && chown -R app /app
USER app

ENTRYPOINT [ "./confluent-gateway" ]