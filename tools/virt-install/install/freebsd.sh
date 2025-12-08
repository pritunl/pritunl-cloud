#!/bin/bash
set -ev
NAME="freebsd"
ISO_URL="https://download.freebsd.org/releases/amd64/amd64/ISO-IMAGES/15.0/FreeBSD-15.0-RELEASE-amd64-dvd1.iso"
ISO_HASH="8cf8e03d8df16401fd5a507480a3270091aa30b59ecf79a9989f102338e359aa"

sudo mkdir -p /var/lib/virt/iso
sudo mkdir -p /var/lib/virt/ks
sudo mkdir -p /var/lib/virt/images

sudo virsh destroy ${NAME} || true
sudo virsh undefine ${NAME} --nvram || true
sudo rm -f /var/lib/virt/${NAME}.qcow2

if [ ! -f "/var/lib/virt/iso/$(basename ${ISO_URL})" ]; then
  sudo wget -P /var/lib/virt/iso ${ISO_URL}
fi

echo "${ISO_HASH} /var/lib/virt/iso/$(basename ${ISO_URL})" | sha256sum --check
if [ $? -ne 0 ]; then
  echo "Checksum for ISO failed"
  exit 1
fi

# Console type [vt100]: Enter
# Welcome: Enter
# Keymap Selection: Enter
# Set Hostname: cloud
# Select Installation Type: Distribution Sets
# Distribution Select: Enter
# Partitioning: Auto (UFS)
# Partition: Enter
# Partition Scheme: Enter
# Partition Editor: Delete freebsd-swap
# Partition Editor: Enter
# New Password: cloud
# Retype New Password: cloud
# Network Configuration: Manual
# Network Configuration: Enter
# Network Configuration: Enter
# Network Configuration: Enter
# Network Configuration: No
# IPv4 DNS #1: 8.8.8.8
# IPv4 DNS #2: 8.8.4.4
# Time Zone Selector: Enter
# Time Zone Confirmation: Enter
# Time & Date: Skip
# Time & Date: Skip
# System Configuration: +ntpd +ntpd_sync_on_start
# System Hardening: Enter
# Add User Accounts: Enter
# Username: cloud
# Full name: Cloud
# Uid: 1000
# Login group: Enter
# Invite cloud into other groups: wheel
# Login class: Enter
# Shell: tcsh
# Home directory: Enter
# Home directory permissions: Enter
# Use password-based authentication: no
# Lock out the account after creation: Enter
# OK: yes
# Add another user: Enter
# Final Configuration: Exit
# Manual Configuration: Enter
# Complete: Enter

sudo virt-install \
  --name ${NAME} \
  --vcpus 8 \
  --memory 8192 \
  --boot uefi \
  --disk path=/var/lib/virt/${NAME}.qcow2,size=8,format=qcow2,bus=virtio \
  --os-variant freebsd14.0 \
  --network network=default \
  --graphics=none \
  --console pty,target_type=serial \
  --cdrom=/var/lib/virt/iso/$(basename ${ISO_URL})

while ! sudo virsh domstate ${NAME} 2>/dev/null | grep -q "shut off"; do
  sleep 1
done

sudo rm -rf /var/lib/virt/init/${NAME}

echo "Compressing image..."
sudo rm -f /var/lib/virt/images/${NAME}_$(date +%y%m%d).qcow2
sudo qemu-img convert -f qcow2 -O qcow2 -c /var/lib/virt/${NAME}.qcow2 /var/lib/virt/images/${NAME}_$(date +%y%m%d).qcow2
sha256sum /var/lib/virt/images/${NAME}_$(date +%y%m%d).qcow2
