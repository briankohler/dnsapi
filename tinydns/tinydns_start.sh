#!/bin/bash 

if [ "${USE_CONSULFS}" == "true" ]
then
  consulfs --allow-other --uid 2 --gid 2 --root dnsfs/ --perm 0777 consul:8500 /service/tinydns/root &
  sleep 1
else
  ln -fs /data /service/tinydns/root
fi

 
/usr/bin/tinydns -c /etc/tinydns/tinydns.conf -l /dev/stdout -d 5
