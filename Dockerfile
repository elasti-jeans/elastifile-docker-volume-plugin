FROM golang:1.10-alpine as builder
COPY . /go/src/github.com/elastifile/elastifile-docker-volume-provisioner
WORKDIR /go/src/github.com/elastifile/elastifile-docker-volume-provisioner
RUN set -ex \
    && apk add --no-cache --virtual .build-deps \
    gcc libc-dev \
    && go install --ldflags '-extldflags "-static"' \
    && apk del .build-deps
CMD ["/go/bin/elastifile-docker-volume-provisioner"]

FROM alpine
RUN mkdir -p /run/docker/plugins /mnt/state /mnt/volumes
COPY --from=builder /go/bin/elastifile-docker-volume-provisioner .
CMD ["elastifile-docker-volume-provisioner"]
