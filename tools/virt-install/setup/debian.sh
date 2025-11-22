#!/bin/bash
set -ev

if [ $(whoami) != "root" ]; then
  echo "Must be run as root"
  exit 1
fi

#############################################################
# starting debian setup
#############################################################

tee /etc/modprobe.d/floppy-blacklist.conf << EOF
blacklist floppy
EOF

apt update
apt -y upgrade
apt -y autoremove

apt -y install bash-completion qemu-guest-agent cloud-init cloud-initramfs-growroot chrony openssh-server

systemctl daemon-reload
systemctl enable qemu-guest-agent.service
systemctl enable cloud-init-local.service
systemctl enable cloud-init.service
systemctl enable cloud-config.service
systemctl enable cloud-final.service

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

passwd -d root
passwd -l root
passwd -d cloud
passwd -l cloud
mkdir -p /home/cloud/.ssh
chown cloud:cloud /home/cloud/.ssh
touch /home/cloud/.ssh/authorized_keys
chown cloud:cloud /home/cloud/.ssh/authorized_keys
chmod 700 /home/cloud/.ssh
chmod 600 /home/cloud/.ssh/authorized_keys
sed -i '/^%sudo/d' /etc/sudoers
tee -a /etc/sudoers << EOF
%sudo ALL=(ALL) NOPASSWD:ALL
EOF

systemctl enable ssh.service
ufw disable

apt clean

rm -f /etc/cloud/cloud.cfg.d/90-installer-network.cfg
rm -f /etc/cloud/cloud.cfg.d/99-installer.cfg
cloud-init clean --machine-id
rm -rf /etc/NetworkManager/system-connections/*

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

#############################################################
# finished debian setup, clear history and shutdown:
# unset HISTFILE && history -c && sudo poweroff
#############################################################
