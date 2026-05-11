FROM gcr.io/distroless/static-debian12
COPY build/bin/store /usr/local/bin/store
ENTRYPOINT ["/usr/local/bin/store"]
CMD ["start", "/var/www/store/api.socket", "/var/www/store/secret.yaml"]
