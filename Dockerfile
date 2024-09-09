FROM golang:1.22 as build

COPY src/ /src
WORKDIR /src
RUN go mod download
RUN CGO_ENABLED=0 go build github.com/dfds/confluent-gateway/cmd/main

FROM alpine

# ADD Curl
RUN apk update && apk add curl && apk add ca-certificates && rm -rf /var/cache/apk/*
# AWS RDS Certificate
RUN curl -o /tmp/rds-combined-ca-bundle.pem https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem  \
  && mv /tmp/rds-combined-ca-bundle.pem /usr/local/share/ca-certificates/rds-combined-ca-bundle.crt \
  && update-ca-certificates

# create and use non-root user
RUN adduser \
  --disabled-password \
  --home /app \
  --gecos '' app \
  && chown -R app /app

USER app

WORKDIR /app
COPY --from=build /src/main /app/confluent-gateway

ENTRYPOINT [ "/app/confluent-gateway" ]