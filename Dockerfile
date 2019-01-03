FROM alpine as builder
RUN apk add --no-cache ca-certificates
USER nobody

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY security-validator /security-validator
# USER nobody
ENTRYPOINT ["/security-validator"]