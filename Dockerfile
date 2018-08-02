FROM golang
RUN XC_OS="linux darwin windows" XC_ARCH="amd64" make bin
COPY pkg/linux_amd64/alicloud_redis_exporter /bin/alicloud_redis_exporter
EXPOSE 9456
CMD /bin/alicloud_redis_exporter