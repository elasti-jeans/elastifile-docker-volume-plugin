FROM golang:1.10-alpine as builder
COPY . /go/src/github.com/elastifile/docker-volume-elastifile
WORKDIR /go/src/github.com/elastifile/docker-volume-elastifile
RUN set -ex \
#    && export GOPATH=$GOPATH:/go/src/github.com/elastifile/docker-volume-elastifile:/go/src/github.com/elastifile/docker-volume-elastifile/vendor/github.com/elastifile/emanage-go \
    && apk add --no-cache --virtual .build-deps \
    gcc libc-dev \
    && go install --ldflags '-extldflags "-static"' \
    && apk del .build-deps
CMD ["/go/bin/docker-volume-elastifile"]

FROM alpine
RUN mkdir -p /run/docker/plugins /mnt/state /mnt/volumes
COPY --from=builder /go/bin/docker-volume-elastifile .
CMD ["docker-volume-elastifile"]
