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

yum -y install sshpass

echo "Welcome, please define how many nodes you wish to find."

echo -n "Number of nodes: "
read -r numnodes

echo "Please specify your IP range that your BMCs are configured on:"

echo -n "BMC range: "
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

echo "Finally, please specify your IPMI password for these nodes:"

echo -n "IPMI Password: "
read -r ipmipass

echo
echo "Answers:"
echo "1. $numnodes"
echo "2. $ipmiaddr"
echo "3. $nodestart"
echo "4. $ipmipass"

echo $ipmipass > ipmipass

tempnodes=$(expr $nodestart + $numnodes)

nodeend=$(expr $tempnodes - 1)

for i in $(seq $nodestart $nodeend); do ssh-keyscan -H $ipmiaddr.$i >> ~/.ssh/known_hosts & done
wait
for i in $(seq $nodestart $nodeend); do
echo $i
nodetmp=$(expr $namestart - 1)
nodeid=$(expr $nodetmp + $i)
nodemac=`sshpass -f ipmipass ssh -l admin ${ipmiaddr}.$i 'adapter -show slot-4 ports' | grep -A5 "21:00:00" | tail -1 | awk '{print $NF}' | tr '[:upper:]' '[:lower:]' | sed 's/\(\w\w\)\(\w\w\)\(\w\w\)\(\w\w\)\(\w\w\)\(\w\w\)/\1:\2:\3:\4:\5:\6/g'`
echo "node $i found"
cd ../baremetal-operator/cmd/make-bm-worker
go run main.go -address ipmi://${ipmiaddr}.$i -password $ipmipass -user admin -disableCertificateVerification -boot-mac ${nodemac} ${nodename}${nodeid} >> /tmp/nodelist
cd -
done
#rm /tmp/gorun
cat /tmp/nodelist | kubectl -n metal-nodes apply -f -
