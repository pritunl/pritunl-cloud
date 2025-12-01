#!/bin/bash
set -ev

if [ $(whoami) != "root" ]; then
  echo "Must be run as root"
  exit 1
fi

#############################################################
# starting fedora setup
#############################################################

tee /etc/modprobe.d/floppy-blacklist.conf << EOF
blacklist floppy
EOF

dnf clean all
dnf -y update
dnf -y install bash-completion qemu-guest-agent dnf-utils cloud-init cloud-utils-growpart chrony openssh-server
dnf -y update
dnf -y remove cockpit-ws

sed -i 's/^GRUB_TIMEOUT=.*/GRUB_TIMEOUT=0/g' /etc/default/grub
grub2-mkconfig -o /boot/grub2/grub.cfg

# cloud-init fix
if [[ "$(cloud-init --version 2>&1)" == *"25.2"* ]]; then
  wget https://dl.fedoraproject.org/pub/fedora/linux/development/rawhide/Everything/x86_64/os/Packages/c/cloud-init-25.3-1.fc44.noarch.rpm
  echo "51c33d2b8780480fa1aa482824e7850e21363e2713a8197e2ed97a8d28c14adb  cloud-init-25.3-1.fc44.noarch.rpm" | dnf -y install cloud-init-25.3-1.fc44.noarch.rpm
  rm -f cloud-init-25.3-1.fc44.noarch.rpm
fi

systemctl daemon-reload
systemctl enable qemu-guest-agent.service
systemctl enable cloud-init-local.service
systemctl enable cloud-init-main.service
systemctl enable cloud-config.service
systemctl enable cloud-final.service

sed -i 's/^installonly_limit=.*/installonly_limit=2/g' /etc/dnf/dnf.conf
sed -i 's/^SELINUX=.*/SELINUX=enforcing/g' /etc/selinux/config || true

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

useradd -m -G adm,video,wheel,systemd-journal cloud || true
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
rm -rf /var/cache/dnf

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
# finished fedora setup, clear history and shutdown:
# unset HISTFILE && history -c && poweroff
#############################################################
