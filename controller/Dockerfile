FROM alpine:latest
RUN apk add --update git ca-certificates && \
    rm -rf /var/cache/apk/*
ADD static /static
ADD controller /bin/controller
COPY ../config/domain.crt /etc/distributionweb/certs/domain.crt
EXPOSE 8080
ENTRYPOINT ["/bin/controller"]
