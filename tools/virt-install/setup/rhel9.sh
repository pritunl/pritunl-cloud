#!/bin/bash
set -e

if [ $(whoami) != "root" ]; then
  echo "Must be run as root"
  exit
fi

echo '#############################################################'
echo 'starting rhel9 setup'
echo '#############################################################'

tee /etc/modprobe.d/floppy-blacklist.conf << EOF
blacklist floppy
EOF

truncate -s 0 /etc/yum/vars/ociregion || true

dnf clean all
dnf -y update
dnf -y install yum-utils
dnf -y install bash-completion qemu-guest-agent yum-utils cloud-init cloud-utils-growpart chrony openssh-server
dnf -y remove cockpit-ws

sed -i 's/^GRUB_TIMEOUT=.*/GRUB_TIMEOUT=0/g' /etc/default/grub
grub2-mkconfig -o /boot/grub2/grub.cfg

systemctl daemon-reload
systemctl enable qemu-guest-agent.service

sed -i 's/^installonly_limit=.*/installonly_limit=2/g' /etc/yum.conf
sed -i 's/^SELINUX=.*/SELINUX=enforcing/g' /etc/selinux/config || true
sed -i 's/^SELINUX=.*/SELINUX=enforcing/g' /etc/sysconfig/selinux || true

sed -i '/^PermitRootLogin/d' /etc/ssh/sshd_config
sed -i '/^PasswordAuthentication/d' /etc/ssh/sshd_config
sed -i '/^ChallengeResponseAuthentication/d' /etc/ssh/sshd_config
sed -i '/^KbdInteractiveAuthentication/d' /etc/ssh/sshd_config
sed -i '/^TrustedUserCAKeys/d' /etc/ssh/sshd_config
sed -i '/^AuthorizedPrincipalsFile/d' /etc/ssh/sshd_config
tee -a /etc/ssh/sshd_config << EOF

PermitRootLogin no
PasswordAuthentication no
ChallengeResponseAuthentication no
KbdInteractiveAuthentication no
TrustedUserCAKeys /etc/ssh/trusted
AuthorizedPrincipalsFile /etc/ssh/principals
EOF
touch /etc/ssh/trusted
touch /etc/ssh/principals
restorecon -v /etc/ssh/trusted
restorecon -v /etc/ssh/principals

useradd -G adm,video,wheel,systemd-journal cloud || true
passwd -d root
passwd -l root
passwd -d cloud
passwd -l cloud
mkdir -p /home/cloud/.ssh
chown cloud:cloud /home/cloud/.ssh
restorecon -v /home/cloud/.ssh
touch /home/cloud/.ssh/authorized_keys
chown cloud:cloud /home/cloud/.ssh/authorized_keys
restorecon -v /home/cloud/.ssh/authorized_keys
chmod 700 /home/cloud/.ssh
chmod 600 /home/cloud/.ssh/authorized_keys
sed -i '/^%wheel/d' /etc/sudoers
tee -a /etc/sudoers << EOF
%wheel ALL=(ALL) NOPASSWD:ALL
EOF

systemctl enable sshd
systemctl restart sshd
systemctl disable firewalld
systemctl stop firewalld
systemctl start chronyd
systemctl enable chronyd

dnf clean all
rm -rf /var/cache/yum
dnf -y update

cloud-init clean --machine-id
rm -rf /etc/NetworkManager/system-connections/*
tee /etc/resolv.conf << EOF
nameserver 8.8.8.8
nameserver 8.8.4.4
EOF

sync
sleep 1

find /var/log -mtime -1 -type f -exec truncate -s 0 {} \;
rm -rf /var/tmp/dnf-*
rm -rf /home/cloud/.cache
shred -u /etc/ssh/*_key /etc/ssh/*_key.pub || true
shred -u /root/.ssh/authorized_keys || true
shred -u /root/.bash_history || true
shred -u /home/cloud/.bash_history || true
shred -u /var/log/lastlog || true
shred -u /var/log/secure || true
shred -u /var/log/utmp || true
shred -u /var/log/wtmp || true
shred -u /var/log/btmp || true
shred -u /var/log/dmesg || true
shred -u /var/log/dmesg.old || true
shred -u /var/lib/systemd/random-seed || true
rm -rf /var/log/*.gz
rm -rf /var/log/*.[0-9]
rm -rf /var/log/*-????????
rm -rf /var/lib/cloud/instances/*
rm -f /var/lib/systemd/random-seed
rm -f /etc/machine-id
touch /etc/machine-id
sync
fstrim -av
sync

echo '#############################################################'
echo 'finished rhel9 setup, clear history and shutdown:'
echo 'unset HISTFILE && history -c && poweroff'
echo '#############################################################'
