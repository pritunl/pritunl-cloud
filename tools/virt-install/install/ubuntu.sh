#!/bin/bash
set -e
export NAME="ubuntu2404"
export ISO_URL="https://releases.ubuntu.com/24.04.2/ubuntu-24.04.2-live-server-amd64.iso"
export ISO_HASH="d6dab0c3a657988501b4bd76f1297c053df710e06e0c3aece60dead24f270b4d"

sudo mkdir -p /var/lib/virt/iso
sudo mkdir -p /var/lib/virt/ks
sudo mkdir -p /var/lib/virt/images
sudo mkdir -p /var/lib/virt/init

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

sudo mkdir /var/lib/virt/init/${NAME}

sudo tee /var/lib/virt/init/${NAME}/user-data << EOF
#cloud-config
autoinstall:
  version: 1
  timezone: "Etc/UTC"
  identity:
    realname: "Cloud"
    username: cloud
    password: "\$6\$x7YEknTyUuNSTTVK\$nq4xoSTrYp7a/Kb1EvtpH97GxG02CFBqELznybQv4XrA7sskq9PI0Y5KADhp9KiwVdrwR6v2IP6wqoxyXj4SP/"
    hostname: cloud
  storage:
    layout:
      name: direct
    config:
      - type: disk
        id: disk0
        match:
          size: largest
      - type: partition
        id: efi-partition
        device: disk0
        size: 100M
        flag: boot
        grub_device: true
      - type: partition
        id: root-partition
        device: disk0
        size: -1
      - type: format
        id: efi-format
        volume: efi-partition
        fstype: fat32
      - type: format
        id: root-format
        volume: root-partition
        fstype: xfs
      - type: mount
        id: efi-mount
        device: efi-format
        path: /boot/efi
      - type: mount
        id: root-mount
        device: root-format
        path: /
EOF

sudo tee /var/lib/virt/init/${NAME}/meta-data << EOF
instance-id: ${NAME}
local-hostname: cloud
EOF

sudo rm -f /var/lib/virt/init/${NAME}.iso
sudo genisoimage -output /var/lib/virt/init/${NAME}.iso \
  -volid cidata \
  -joliet \
  -rock \
  -input-charset utf-8 \
  /var/lib/virt/init/${NAME}/user-data \
  /var/lib/virt/init/${NAME}/meta-data

sudo rm -rf /var/lib/virt/init/${NAME}

sudo virt-install \
  --name ${NAME} \
  --vcpus 8 \
  --memory 8192 \
  --boot uefi \
  --disk path=/var/lib/virt/${NAME}.qcow2,size=8,format=qcow2,bus=virtio \
  --disk path=/var/lib/virt/init/${NAME}.iso,device=cdrom \
  --os-variant ubuntu-lts-latest \
  --network network=default \
  --graphics=none \
  --console pty,target_type=serial \
  --location=/var/lib/virt/iso/$(basename ${ISO_URL}),kernel=casper/hwe-vmlinuz,initrd=casper/hwe-initrd \
  --extra-args="console=ttyS0 serial autoinstall"

while ! sudo virsh domstate ${NAME} 2>/dev/null | grep -q "shut off"; do
  sleep 1
done

echo "Compressing image..."

sudo rm -f /var/lib/virt/images/${NAME}_$(date +%y%m).qcow2
sudo qemu-img convert -f qcow2 -O qcow2 -c /var/lib/virt/${NAME}.qcow2 /var/lib/virt/images/${NAME}_$(date +%y%m).qcow2
sha256sum /var/lib/virt/images/${NAME}_$(date +%y%m).qcow2
