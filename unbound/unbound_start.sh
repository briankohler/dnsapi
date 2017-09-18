#!/bin/bash 

unbound-control-setup > /dev/null
cp /etc/unbound/unbound_server.* /
export CONSUL_IP=$(ping -c 1 consul | head -n1 | cut -d" " -f3 | sed 's/[\(\)]//g' | awk -F: '{print $1}')
export TINYDNS_IP=$(ping -c 1 tinydns | head -n1 | cut -d" " -f3 | sed 's/[\(\)]//g' | awk -F: '{print $1}')
cp /etc/unbound.conf /etc/unbound/unbound.conf
sed -i "s/{AUTH_DOMAINS}/$AUTH_DOMAINS/g" /etc/unbound/unbound.conf
sed -i "s/{CONSUL_IP}/$CONSUL_IP/g" /etc/unbound/unbound.conf
sed -i "s/{TINYDNS_IP}/$TINYDNS_IP/g" /etc/unbound/unbound.conf
unbound -c /etc/unbound/unbound.conf -d -v
