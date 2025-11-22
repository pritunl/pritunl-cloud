#!/bin/bash
set -ev

if [ $(whoami) != "root" ]; then
  echo "Must be run as root"
  exit 1
fi

#############################################################
# starting freebsd setup
#############################################################

env PAGER=/bin/cat freebsd-update fetch
freebsd-update install || true
pkg update
pkg upgrade -y

sysrc -f /boot/loader.conf autoboot_delay=0

pkg search cloud-init
pkg install -y dual-dhclient py311-cloud-init
sysrc dhclient_program="/usr/local/sbin/dual-dhclient"

pw mod user root -w no
pw mod user cloud -w no

mkdir -p /home/cloud/.ssh
chown cloud:cloud /home/cloud/.ssh
touch /home/cloud/.ssh/authorized_keys
chown cloud:cloud /home/cloud/.ssh/authorized_keys
chmod 700 /home/cloud/.ssh
chmod 600 /home/cloud/.ssh/authorized_keys

sed -i "" '/^PermitRootLogin/d' /etc/ssh/sshd_config
sed -i "" '/^PasswordAuthentication/d' /etc/ssh/sshd_config
sed -i "" '/^ChallengeResponseAuthentication/d' /etc/ssh/sshd_config
sed -i "" '/^KbdInteractiveAuthentication/d' /etc/ssh/sshd_config
sed -i "" '/^TrustedUserCAKeys/d' /etc/ssh/sshd_config
sed -i "" '/^AuthorizedPrincipalsFile/d' /etc/ssh/sshd_config
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

tee /etc/sudoers << EOF
%wheel ALL=(ALL) NOPASSWD:ALL
EOF
chmod 600 /etc/sudoers

sysrc swapoff="YES"
sysrc ifconfig_vtnet0=""
sysrc cloudinit_enable="YES"

tee /usr/local/etc/cloud/cloud.cfg.d/99_cloud.cfg << EOF
datasource_list: [ NoCloud ]
EOF

cloud-init clean --machine-id

tee /usr/local/etc/rc.d/cloudinitfix << EOF
#!/bin/sh

# PROVIDE: cloudinitfix
# REQUIRE: FILESYSTEMS NETWORKING ldconfig devd
# BEFORE:  LOGIN cloudconfig cloudinit

. /etc/rc.subr

PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

name="cloudinitfix"
start_cmd="cloudinitfix_start"
stop_cmd=":"
rcvar="cloudinit_enable"

cloudinitfix_start()
{
  rm -rf /var/lib/cloud/instances
}

load_rc_config \$name

: \${cloudinitfix_enable="NO"}

run_rc_command "\$1"
EOF
chmod 755 /usr/local/etc/rc.d/cloudinitfix

tee /etc/resolv.conf << EOF
nameserver 8.8.8.8
nameserver 8.8.4.4
EOF
rm -f /etc/resolv.conf.bak

sync
rm -f /var/db/dhclient.leases.vtnet0
rm -f /var/db/dhclient6.leases
rm -f setup.sh
rm -rf /root/.cache
rm -rf /home/cloud/.cache
rm -f /etc/ssh/*_key*
rm -f /root/.ssh/authorized_keys
rm -f /root/.history
rm -f /home/cloud/.history
rm -f /root/.bash_history
rm -f /home/cloud/.bash_history
find /var/log -mtime -1 -type f -exec truncate -s 0 {} \;
sync

#############################################################
# finished freebsd setup, clear history and shutdown:
# unset history && unset HISTFILE && poweroff
#############################################################
