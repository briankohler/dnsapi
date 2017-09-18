FROM alpine:latest

ADD unbound.conf /etc/unbound.conf 
RUN apk add --no-cache ca-certificates unbound net-tools openssl bash 
VOLUME /etc/unbound
ADD unbound_start.sh /start.sh
EXPOSE 53/udp
CMD /./start.sh


