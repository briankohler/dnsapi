#!/bin/bash 

while [ 1 ]
do
  if [ -f /usr/bin/consul ]
  then
    /bin/cp /usr/bin/consul /usr/local/bin/
    /usr/local/bin/consul watch -http-addr=${CONSUL_HOST}:${CONSUL_PORT} -type=keyprefix -prefix=${CONSUL_KEYSPACE}/ 'sleep 2 && /usr/local/bin/dnsapidumper' > /dev/stdout 2>&1
  else
    sleep $POLLING_INTERVAL
    /usr/local/bin/dnsapidumper
  fi
done 
