#!/bin/bash
# Copyright 2023 Logicalis UKI. All Rights Reserved.
#
# Licensed under the Functional Source License, Version 1.0, Apache 2.0 Change License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/logicalisuki/ubiquity/blob/main/LICENSE.md
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# It also allows for the transition of this software to an Apache 2.0 Licence
# on the second anniversary of the date we make the software available.
# See the License for the specific language governing permissions and
# limitations under the License.

## Filename
echo "Welcome, please define the filename you would like your nodelist to be saved in"

echo -n "Nodelist name: "
read -r nodelistname

## Number of nodes
echo "Please define how many nodes you wish to define."

echo -n "Number of nodes: "
read -r numnodes

## IP range
echo "Please specify your IP range that your BMCs are configured on, ending with 0"

echo -n "BMC range (ex 192.168.1.0):"
read -r ipmiaddr

echo "Please specify your starting node (1-$numnodes):"

echo -n "Node start: "
read -r nodestart

echo "Please specify your node naming convention:"
echo -n "Name: "
read -r nodename

echo "Please specify your starting node number:"
echo -n "NamingStart: "
read -r namestart

echo "Please specify your IPMI details for these nodes:"

echo -n "IPMI User: "
read -r ipmiuser
echo -n "IPMI Password: "
read -r ipmipass

echo "Please specify your PXE boot device for these nodes (in the format of Function:Slot:Port):"

echo -n "PXE Boot Device: "
read -r pxedev

echo "Please specify the user information you wish to be provisioned on the nodes:"

echo -n "SSH user: "
read -r username
echo -n "SSH password for $username: "
read -r password

echo "Please specify the first DNS server you wish these nodes to use:"
dns1prop=0.0.0.0
echo -n "DNS server 1 (leave for blank): "
read -r dns1
if [[ -z $dns1 ]]; then
dns1=${dns1prop}
fi

echo "Please specify the second first DNS server you wish these nodes to use:"
dns2prop=0.0.0.0
echo -n "DNS server 2 (leave for blank): "
read -r dns2
if [[ -z $dns2 ]]; then
dns2=${dns2prop}
fi

echo "Please specify the management network interface you wish these nodes to use:"

echo -n "Management network interface: "
read -r net1

echo "Please specify the management network IP range you wish these nodes to use:"

echo -n "Management network IP range: "
read -r ipaddr1

echo "Please specify the management network IP netmask you wish these nodes to use:"

echo -n "Management network IP netmask: "
read -r netmask1

echo "Please specify the management network name you wish these nodes to use:"
netname1prop="mgmtnet"
echo -n "Management network name ($netname1prop): "
read -r netname1

if [[ -z $netname1 ]]; then
netname1=${netname1prop}
fi

echo "Please specify the management network gateway IP you wish these nodes to use:"
gateway1prop="10.212.87.254"
echo -n "Management network gateway (ex. $gateway1prop): "
read -r gateway1

if [[ -z $gateway1 ]]; then
gateway1=${gateway1prop}
fi

echo "Please specify the high-speed fabric network interface you wish these nodes to use:"

echo -n "High-speed fabric network interface: "
read -r net2

echo "Please specify the high-speed fabric network IP range you wish these nodes to use:"

echo -n "High-speed fabric network IP range: "
read -r ipaddr2

echo "Please specify the high-speed fabric network IP netmask you wish these nodes to use:"

echo -n "High-speed fabric network IP netmask: "
read -r netmask2

echo "Please specify the high-speed fabric network name you wish these nodes to use:"
netname2prop="hfinet"
echo -n "Management network name ($netname2prop): "
read -r netname2

if [[ -z $netname2 ]]; then
netname2=${netname2prop}
fi

echo "Please specify the high-speed fabric network gateway IP you wish these nodes to use:"
gateway2prop="0.0.0.0"
echo -n "Management network gateway ($gateway2prop): "
read -r gateway2

if [[ -z $gateway2 ]]; then
gateway2=${gateway2prop}
fi

echo
echo "Answers:"
echo "1. $numnodes"
echo "2. $ipmiaddr"
echo "3. $nodestart"
echo "4. $ipmiuser"
echo "5. $ipmipass"
echo "6. $pxedev"
echo "7. $username"
echo "8. $password"
echo "9. $dns1"
echo "10. $dns2"
echo "11. $net1"
echo "12. $ipaddr1"
echo "13. $netmask1"
echo "13. $netname1"
echo "14. $gateway1"
echo "15. $net2"
echo "16. $ipaddr2"
echo "17. $netmask2"
echo "18. $netname2"
echo "19. $gateway2"

echo "#name,address,ipmiuser,ipmipass,pxedev,user,password,mac,dns1,dns2,net1,ip1,netmask1,netname1,gateway1,net2,ip2,netmask2,netname2,gateway2" > $nodelistname

ipmi1=${ipmiaddr}.${nodestart}

ip1=${ipaddr1}.${nodestart}

ip2=${ipaddr2}.${nodestart}

num=$numnodes

ipmiarr=($(echo $ipmi1|sed  's/\./ /g'))
ip1arr=($(echo $ip1|sed  's/\./ /g'))

if [[ ip2 != ".${nodestart}" ]]; then
ip2arr=($(echo $ip2|sed  's/\./ /g'))
fi

ipmiip=$((ipmiarr[0] << 24))
ipmiip=$((ipmiip | ipmiarr[1] << 16))
ipmiip=$((ipmiip | ipmiarr[2] << 8))
ipmiip=$((ipmiip | ipmiarr[3]))

if [[ ! -z ip1arr ]]; then
n1ip=$((ip1arr[0] << 24))
n1ip=$((n1ip | ip1arr[1] << 16))
n1ip=$((n1ip | ip1arr[2] << 8))
n1ip=$((n1ip | ip1arr[3]))
fi

if [[ ! -z ip2arr ]]; then
n2ip=$((ip2arr[0] << 24))
n2ip=$((n2ip | ip2arr[1] << 16))
n2ip=$((n2ip | ip2arr[2] << 8))
n2ip=$((n2ip | ip2arr[3]))
fi

lnip=$((ipmiip + num))
nodetmp=$(expr $namestart - 1)

while [ $ipmiip -lt $lnip ]; do
    let $((ipmiip++))
    let $((nodetmp++))

    ipmiarr[0]=$((ipmiip >> 24))
    ipmiarr[1]=$(((ipmiip >> 16) & 255))
    ipmiarr[2]=$(((ipmiip >> 8) & 255))
    ipmiarr[3]=$((ipmiip & 255))

    if [[ ! -z ip1arr ]]; then
    let $((n1ip++))
    ip1arr[0]=$((n1ip >> 24))
    ip1arr[1]=$(((n1ip >> 16) & 255))
    ip1arr[2]=$(((n1ip >> 8) & 255))
    ip1arr[3]=$((n1ip & 255))
    fi

    if [[ ! -z ip2arr ]]; then
    let $((n2ip++))
    ip2arr[0]=$((n2ip >> 24))
    ip2arr[1]=$(((n2ip >> 16) & 255))
    ip2arr[2]=$(((n2ip >> 8) & 255))
    ip2arr[3]=$((n2ip & 255))
    fi

    if [[ -z ip2arr ]]; then
    printf "%s%s,%d.%d.%d.%d,%s,%s,%s,%s,%s,%s%s-mac,%s,%s,%s,%d.%d.%d.%d,%s,%s,%s\n" ${nodename} ${nodetmp} ${ipmiarr[0]} ${ipmiarr[1]} ${ipmiarr[2]} ${ipmiarr[3]} ${ipmiuser} ${ipmipass} ${pxedev} ${username} ${password} ${nodename} ${nodetmp} ${dns1} ${dns2} ${net1} ${ip1arr[0]} ${ip1arr[1]} ${ip1arr[2]} ${ip1arr[3]} ${netmask1} ${netname1} ${gateway1} >> $nodelistname
    fi

    if [[ ! -z ip2arr ]]; then
    printf "%s%s,%d.%d.%d.%d,%s,%s,%s,%s,%s,%s%s-mac,%s,%s,%s,%d.%d.%d.%d,%s,%s,%s,%s,%d.%d.%d.%d,%s,%s,%s\n" ${nodename} ${nodetmp} ${ipmiarr[0]} ${ipmiarr[1]} ${ipmiarr[2]} ${ipmiarr[3]} ${ipmiuser} ${ipmipass} ${pxedev} ${username} ${password} ${nodename} ${nodetmp} ${dns1} ${dns2} ${net1} ${ip1arr[0]} ${ip1arr[1]} ${ip1arr[2]} ${ip1arr[3]} ${netmask1} ${netname1} ${gateway1} ${net2} ${ip2arr[0]} ${ip2arr[1]} ${ip2arr[2]} ${ip2arr[3]} ${netmask2} ${netname2} ${gateway2} >> $nodelistname
    fi

done
