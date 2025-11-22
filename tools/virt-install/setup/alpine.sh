#!/bin/sh
set -ev

if [ $(whoami) != "root" ]; then
  echo "Must be run as root"
  exit 1
fi

#############################################################
# starting alpine setup
#############################################################

tee /etc/motd << EOF
Welcome to Alpine!

The Alpine Wiki contains a large amount of how-to guides and general
information about administrating Alpine systems.
See <https://wiki.alpinelinux.org/>.

EOF

echo "iso9660" > /etc/filesystems
sed -i 's/^GRUB_TIMEOUT=.*/GRUB_TIMEOUT=0/g' /etc/default/grub
grub-mkconfig -o /boot/grub/grub.cfg

rc-update add sshd default
rc-update add chronyd default
rc-update add qemu-guest-agent default
setup-cloud-init

tee /etc/init.d/cloud-fix << EOF
#!/sbin/openrc-run

description="cloud-init final fix stage"

depend() {
  after cloud-config
  provide cloud-fix
}

start() {
  if grep -q 'cloud-init=disabled' /proc/cmdline; then
    ewarn "\$RC_SVCNAME is disabled via /proc/cmdline."
  elif test -e /etc/cloud/cloud-init.disabled; then
    ewarn "\$RC_SVCNAME is disabled via cloud-init.disabled file"
  else
    ebegin "cloud-init fix"
    cloud-init modules --mode final
    eend \$?
  fi
}
EOF
chmod +x /etc/init.d/cloud-fix
rc-update add cloud-fix default

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

mkdir -p /home/alpine
chown alpine:alpine /home/alpine
chmod 700 /home/alpine
usermod -l cloud alpine
usermod -m -d /home/cloud cloud
groupmod -n cloud alpine
usermod -aG adm,wheel cloud
passwd -d root
passwd -l root
passwd -d cloud
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

cloud_password=$(tr -dc 'A-Za-z0-9!@#$%^&*()_+~' < /dev/urandom | head -c 64)
echo "cloud:$cloud_password" | chpasswd
passwd -u cloud

cloud-init clean --machine-id
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
shred -u /root/.ash_history || true
shred -u /home/cloud/.ash_history || true
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
sync
fstrim -v /
sync

#############################################################
# finished alpine setup, clear history and shutdown:
# unset HISTFILE && poweroff
#############################################################
