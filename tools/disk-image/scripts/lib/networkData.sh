#!/bin/bash
# Copyright 2023 Logicalis UKI. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0. Previously licensed under the Functional Source License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/logicalisuki/ubiquity-open/blob/main/LICENSE
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# This software was previously licensed under the Functional Source License but has now transitioned to an Apache 2.0 License
# as of June 2025.
# See the License for the specific language governing permissions and
# limitations under the License.

# Define function to build JSON output
build_json_output() {
  hostname="$1"
  dns1="$2"
  dns2="$3"
  interface_names=("${!4}")
  ip_addresses=("${!5}")
  netmasks=("${!6}")
  network_names=("${!7}")
  gateways=("${!8}")

  echo '{
  "services": [
    {
      "type": "dns",
      "address": "'"$dns1"'"
    },
    {
      "type": "dns",
      "address": "'"$dns2"'"
    }
  ],
  "networks": ['

  for (( i=0; i<${#interface_names[@]}; i++ )); do
    echo '    {
      "network_id": "network'"$i"'",
      "link": "'${interface_names[$i]}'",
      "type": "ipv4",'
    if [ "${network_names[$i]}" != "" ]; then
      echo '      "id": "'${network_names[$i]}'",'
    fi
    if [ "${gateways[$i]}" != "0.0.0.0" ]; then
      echo '      "gateway": "'${gateways[$i]}'"',
    fi
    if [ "${ip_addresses[$i]}" != "" ]; then
      echo '      "ip_address": "'${ip_addresses[$i]}'",'
    fi
    if [ "${netmasks[$i]}" != "" ]; then
      echo '      "netmask": "'${netmasks[$i]}'",'
    fi
    echo '      "routes": []
    }'
    if [ $i -ne $((${#interface_names[@]}-1)) ]; then
      echo '    ,'
    fi
  done

  echo '  ],
  "links": ['

  for (( i=0; i<${#interface_names[@]}; i++ )); do
    echo '    {
      "type": "phy",
      "id": "'${interface_names[$i]}'",
      "name": "'${interface_names[$i]}'",'
    if [[ ${interface_names[$i]} == ib* ]]; then
      echo '      "mtu": 2044'
    else
      echo '      "mtu": 1500'
    fi
    echo '    }'
    if [ $i -ne $((${#interface_names[@]}-1)) ]; then
      echo '    ,'
    fi
  done

  echo '  ]
}'
}

# Check if arguments were provided
if [ $# -eq 0 ]; then
  echo "Usage: $0 -hostname <hostname> -dns1 <dns1> -dns2 <dns2> -net1 <interface_name> <ip_address> <netmask> <network_name> <gateway> [-net2 <interface_name> <ip_address> <netmask> <network_name> <gateway>] ..."
  exit 1
fi

# Parse arguments
while [ $# -gt 0 ]; do
  case "$1" in
    -hostname)
      hostname="$2"
      shift 2
      ;;
    -dns1)
      dns1="$2"
      shift 2
      ;;
    -dns2)
      dns2="$2"
      shift 2
      ;;
    -net*)
      net_num=$(echo $1 | sed 's/-net//')
      if [ "$net_num" -gt 0 ]; then
        eval "interface_name_${net_num}=("$2")"
        eval "ip_address_${net_num}=("$3")"
        eval "netmask_${net_num}=("$4")"
        eval "network_name_${net_num}=("$5")"
        eval "gateway_${net_num}=("$6")"
        shift 6
        let "numints++"
      else
        echo "Invalid network interface number: $net_num"
        exit 1
      fi
      ;;
    *)
      echo "Invalid argument: $1"
      exit 1
      ;;
  esac
done

# Check if required arguments were provided
if [ -z "$hostname" ] || [ -z "$dns1" ] || [ -z "$dns2" ]; then
  echo "Usage: $0 -hostname <hostname> -dns1 <dns1> -dns2 <dns2> -net1 <interface_name> <ip_address> <netmask> <network_name> <gateway> [-net2 <interface_name> <ip_address> <netmask> <network_name> <gateway>] ..."
  exit 1
fi

# Create arrays for interface variables
interface_names=()
ip_addresses=()
netmasks=()
network_names=()
gateways=()

# Loop through all interface variables and add them to arrays
#for i in $(seq 1 $(( $# / 5 ))); do
#for i in $(seq 1  ); do
for ((i = 1; i <= $numints; i++)); do
  interface_names+=($(eval echo \$interface_name_$i))
  ip_addresses+=($(eval echo \$ip_address_$i))
  netmasks+=($(eval echo \$netmask_$i))
  network_names+=($(eval echo \$network_name_$i))
  gateways+=($(eval echo \$gateway_$i))
done
# Call function to build JSON output
build_json_output "$hostname" "$dns1" "$dns2" interface_names[@] ip_addresses[@] netmasks[@] network_names[@] gateways[@]
