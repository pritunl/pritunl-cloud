#!/bin/bash
set -ev

if [ $(whoami) != "root" ]; then
  echo "Must be run as root"
  exit 1
fi

#############################################################
# starting arch setup
#############################################################

sed -i 's/^timeout.*/timeout 0/' /boot/loader/loader.conf

pacman -Syu
pacman -Sy --noconfirm chrony qemu-guest-agent cloud-init cloud-guest-utils dhcpcd

systemctl enable sshd
systemctl enable chronyd

systemctl enable cloud-init-local
systemctl enable cloud-init-main
systemctl enable cloud-config
systemctl enable cloud-final
systemctl disable systemd-networkd-wait-online
systemctl mask systemd-networkd-wait-online

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

useradd -m -G wheel,systemd-journal cloud
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
sed -i '/^%wheel/d' /etc/sudoers
tee -a /etc/sudoers << EOF
%wheel ALL=(ALL) NOPASSWD:ALL
EOF

cloud-init clean --machine-id
tee /etc/resolv.conf << EOF
nameserver 8.8.8.8
nameserver 8.8.4.4
EOF

sync
sleep 1

find /var/log -mtime -1 -type f -exec truncate -s 0 {} \;
rm -rf /var/tmp/*
rm -rf /home/cloud/.cache
shred -u /etc/ssh/*_key /etc/ssh/*_key.pub 2>/dev/null || true
shred -u /root/.ssh/authorized_keys 2>/dev/null || true
shred -u /root/.bash_history 2>/dev/null || true
shred -u /home/cloud/.bash_history 2>/dev/null || true
shred -u /var/log/lastlog 2>/dev/null || true
shred -u /var/log/secure 2>/dev/null || true
shred -u /var/log/utmp 2>/dev/null || true
shred -u /var/log/wtmp 2>/dev/null || true
shred -u /var/log/btmp 2>/dev/null || true
shred -u /var/log/dmesg 2>/dev/null || true
shred -u /var/log/dmesg.old 2>/dev/null || true
shred -u /var/lib/systemd/random-seed 2>/dev/null || true
rm -rf /var/log/*.gz
rm -rf /var/log/*.[0-9]
rm -rf /var/log/*-????????
rm -rf /var/lib/cloud/instances/*
rm -f /var/lib/systemd/random-seed
sync
fstrim -v /
sync

#############################################################
# finished arch setup, clear history and shutdown:
# unset HISTFILE && poweroff
#############################################################
