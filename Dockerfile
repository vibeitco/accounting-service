FROM alpine:3.7
COPY ./svc ./svc
COPY ./config ./config

RUN apk add --no-cache curl ca-certificates && update-ca-certificates

EXPOSE 8080

CMD ["./svc"]
