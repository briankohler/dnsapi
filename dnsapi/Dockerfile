FROM          golang:alpine as BUILDER

ENV            CONSUL_VERSION 0.9.3
RUN            apk add --no-cache git make unbound wget tar fuse tinydns \
                && cd /tmp \
                && wget https://releases.hashicorp.com/consul/${CONSUL_VERSION}/consul_${CONSUL_VERSION}_linux_amd64.zip?_ga=2.57932269.677152365.1504906949-1004122255.1501406113 -O consul.zip \
                && unzip consul.zip && mv consul /usr/bin/ && rm *.zip
ADD            ./dnsapi /go/src/github.com/briankohler/dnsapi
ADD            ./dnsapidumper /go/src/github.com/briankohler/dnsapidumper
ADD            ./consulhelper /go/src/github.com/briankohler/consulhelper
ADD            ./helper /go/src/github.com/briankohler/helper
ADD            ./log /go/src/github.com/briankohler/log
ADD            ./logmiddleware /go/src/github.com/briankohler/logmiddleware
RUN            go get github.com/briankohler/dnsapi
RUN            go get github.com/briankohler/dnsapidumper
RUN            go get github.com/bwester/consulfs/cmd/consulfs
RUN            go install github.com/briankohler/dnsapi
RUN            go install github.com/briankohler/dnsapidumper 
RUN            go install github.com/bwester/consulfs/cmd/consulfs



FROM          alpine:latest

ENV           DNSAPIDUMPER_TINYDATADIR /service/tinydns/root
ENV           POLLING_INTERVAL 60
RUN           apk add --no-cache py2-pip git curl bash make fuse && mkdir -p /service/tinydns \
                && pip install pip --upgrade && pip install awscli awscli-plugin-endpoint
COPY          --from=builder /go/bin/dnsapi /usr/local/bin/dnsapi
COPY          --from=builder /go/bin/dnsapidumper /usr/local/bin/dnsapidumper
COPY          --from=builder /usr/bin/consul /usr/bin/consul
COPY          --from=builder /usr/sbin/unbound-control /usr/bin/unbound-control
COPY          --from=builder /usr/bin/tinydns-data /usr/bin/tinydns-data
COPY          --from=builder /go/bin/consulfs /usr/bin/consulfs
COPY          run_dumper.sh  /usr/bin/run_dumper.sh
COPY          tinydns_makefile $DNSAPIDUMPER_TINYDATADIR/../Makefile
COPY          dnsapi_start.sh /start.sh 

EXPOSE        9080
CMD           /./start.sh



