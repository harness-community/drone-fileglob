FROM alpine:3.19.4 as builder

RUN apk --no-cache add ca-certificates

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ADD release/linux/amd64/drone-findfiles /bin/

ENTRYPOINT ["/bin/drone-findfiles"]