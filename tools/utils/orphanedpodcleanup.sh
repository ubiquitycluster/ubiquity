#!/bin/bash
while true ; do
for o in $(tail /var/log/messages | grep -o  -E 'orphaned pod \\"((\w|-)+)\\' | cut -d" " -f3 | grep -oE '(\w|-)+' | uniq); do
        p="/var/lib/kubelet/pods/$o/volumes/"
        if [ -d "$p" ] ; then
          echo "Removing $o"
          rm -rf "$p"
        fi
done
for o in $(journalctl -eu k3s-agent | grep -o  -E 'orphaned pod \\"((\w|-)+)\\' | cut -d" " -f3 | grep -oE '(\w|-)+' | uniq); do
        p="/var/lib/kubelet/pods/$o/volumes/"
        if [ -d "$p" ] ; then
          echo "Removing $o"
          rm -rf "$p"
        fi
done
for o in $(journalctl -eu k3s-server | grep -o  -E 'orphaned pod \\"((\w|-)+)\\' | cut -d" " -f3 | grep -oE '(\w|-)+' | uniq); do
        p="/var/lib/kubelet/pods/$o/volumes/"
        if [ -d "$p" ] ; then
          echo "Removing $o"
          rm -rf "$p"
        fi
done
sleep 2
done
