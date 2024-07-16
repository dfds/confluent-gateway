FROM dfdsdk/pgup:v0.0.1

RUN apk update && apk add curl && apk add ca-certificates && rm -rf /var/cache/apk/*

# AWS RDS Certificate
RUN curl -o /tmp/rds-combined-ca-bundle.pem https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem \
    && mv /tmp/rds-combined-ca-bundle.pem /usr/local/share/ca-certificates/rds-combined-ca-bundle.crt \
    && update-ca-certificates

COPY . /db
