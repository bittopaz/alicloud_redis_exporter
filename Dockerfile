FROM golang
COPY pkg/linux_amd64/alicloud_redis_exporter /bin/alicloud_redis_exporter
EXPOSE 9456
CMD /bin/alicloud_redis_exporter