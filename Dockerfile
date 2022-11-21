FROM alpine

WORKDIR /app
COPY ./.output/app ./

USER nonroot:nonroot

ENTRYPOINT [ "./confluent-gateway" ]