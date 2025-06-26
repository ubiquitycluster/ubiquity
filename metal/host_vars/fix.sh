#!/bin/bash

for i in {1..18}
do
echo "  - name: vlan121
    device: eno1np0
    ip: 10.148.121.$(( i + 29 ))
    cidr: 24
    netmask: 255.255.255.0
    gateway: 10.148.121.254
    nameserver: '10.144.1.248'
    vlanid: 121" >> cn$i.yml 
done 
