FROM golang:alpine as BUILDER
RUN apk add --no-cache git fuse
RUN go get github.com/bwester/consulfs/cmd/consulfs
RUN go install github.com/bwester/consulfs/cmd/consulfs

FROM alpine:latest
ENV IP 0.0.0.0
ENV ROOT /service/tinydns/root
ENV GID 2
ENV UID 2
ENV HIDETTL 1
ENV DEBUG_LEVEL 5
ENV DATALIMIT 300000
RUN apk add --no-cache tinydns fuse bash && mkdir -p /service/tinydns
COPY tinydns.conf /etc/tinydns/tinydns.conf
COPY --from=builder /go/bin/consulfs /usr/bin/consulfs
COPY tinydns_start.sh /start.sh
CMD /./start.sh

