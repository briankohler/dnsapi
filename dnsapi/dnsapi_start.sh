#!/bin/bash 

if [ "${USE_CONSULFS}" == "true" ]
then
  consulfs --allow-other --root dnsfs/ --perm 0777 consul:8500 /service/tinydns/root &
  sleep 1
fi

ln -fs /etc/unbound/unbound_control.pem /unbound_control.pem
ln -fs /etc/unbound/unbound_control.key /unbound_control.key
ln -fs /etc/unbound/unbound_server.pem /unbound_server.pem
ln -fs /etc/unbound/unbound_server.key /unbound_server.key
if [ "${DNSAPIDUMPER_TINYDATADIR}" != "/data" ]
then
  ln -fs /data $DNSAPIDUMPER_TINYDATADIR 
fi

/usr/bin/run_dumper.sh &
/usr/local/bin/dnsapi &

sleep 5

/usr/local/bin/dnsapidumper

curl -s -XPOST localhost:9080/v2/dnsapi/NS/${AUTH_DOMAINS}/127.0.0.1/ns.${AUTH_DOMAINS}/300
curl -s -XPOST localhost:9080/v2/dnsapi/NS/10.in-addr.arpa/127.0.0.1/ns.${AUTH_DOMAINS}/300
curl -s -XPOST localhost:9080/v2/dnsapi/SOA/${AUTH_DOMAINS}/ns.${AUTH_DOMAINS}
curl -s -XPOST localhost:9080/v2/dnsapi/SOA/10.in-addr.arpa/ns.${AUTH_DOMAINS}

wait

