FROM golang as builder
RUN XC_OS="linux" XC_ARCH="amd64" make bin
COPY pkg/linux_amd64/alicloud_redis_exporter /bin/alicloud_redis_exporter

FROM alpine:latest
COPY --from=builder /bin/alicloud_redis_exporter .
EXPOSE 9456
CMD ./alicloud_redis_exporter