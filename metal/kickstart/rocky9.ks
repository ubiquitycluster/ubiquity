#version=RHEL9
text
network --bootproto=dhcp --device=link --activate
url --url="http://CHANGEME/rocky/9/BaseOS/x86_64/os/"
lang en_US.UTF-8
keyboard us
timezone UTC
rootpw --iscrypted CHANGEME
user --name=admin --groups=wheel --iscrypted --password=CHANGEME
selinux --enforcing
firewall --enabled --service=ssh
services --enabled=NetworkManager,sshd
skipx
zerombr
clearpart --all --initlabel
autopart --type=lvm
bootloader --location=mbr
%packages
@^server-product-environment
kexec-tools
%end
reboot