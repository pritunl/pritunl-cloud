#!/bin/bash
set -ev
NAME="alpinelinux"
ISO_URL="https://dl-cdn.alpinelinux.org/alpine/v3.22/releases/x86_64/alpine-virt-3.22.0-x86_64.iso"
ISO_HASH="c935c3715c80ca416e2ce912c552f0cbfd8531219b7973ae4a600873c793eb1b"

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
#alpine-config
apk:
  repositories:
    - base_url: https://dl-cdn.alpinelinux.org/alpine
      repos:
        - main
        - community
packages:
  - curl
  - dosfstools
  - grub-efi
  - xfsprogs
  - xfsprogs-extra
  - sudo
  - chrony
  - openssh
  - qemu-guest-agent
  - cloud-init
  - cloud-utils-growpart
runcmd:
  - rm /etc/runlevels/*/tiny-cloud*
  - lbu include /root/.ssh /home/alpine/.ssh
  - ERASE_DISKS=/dev/vda USE_EFI=1 DISKLABEL=gpt ROOTFS=xfs BOOT_SIZE=100 SWAP_SIZE=0 setup-disk -m sys /dev/vda
  - reboot
EOF

sudo tee /var/lib/virt/init/${NAME}/meta-data << EOF
hostname: cloud
EOF

sudo rm -f /var/lib/virt/init/${NAME}.iso
sudo genisoimage -output /var/lib/virt/init/${NAME}.iso \
  -volid cidata \
  -joliet \
  -rock \
  -input-charset utf-8 \
  /var/lib/virt/init/${NAME}/user-data \
  /var/lib/virt/init/${NAME}/meta-data

sudo virt-install \
  --name ${NAME} \
  --vcpus 8 \
  --memory 8192 \
  --boot uefi,firmware.feature0.name=secure-boot,firmware.feature0.enabled=no \
  --disk path=/var/lib/virt/${NAME}.qcow2,size=8,format=qcow2,bus=virtio \
  --os-variant alpinelinux3.20 \
  --network network=default \
  --graphics=none \
  --console pty,target_type=serial \
  --location=/var/lib/virt/iso/$(basename ${ISO_URL}),kernel=boot/vmlinuz-virt,initrd=boot/initramfs-virt \
  --cloud-init meta-data=/var/lib/virt/init/${NAME}/meta-data,user-data=/var/lib/virt/init/${NAME}/user-data \
  --extra-args="console=ttyS0 autoinstall"

while ! sudo virsh domstate ${NAME} 2>/dev/null | grep -q "shut off"; do
  sleep 1
done

sudo rm -rf /var/lib/virt/init/${NAME}
echo "Compressing image..."

sudo rm -f /var/lib/virt/images/${NAME}_$(date +%y%m%d).qcow2
sudo qemu-img convert -f qcow2 -O qcow2 -c /var/lib/virt/${NAME}.qcow2 /var/lib/virt/images/${NAME}_$(date +%y%m%d).qcow2
sha256sum /var/lib/virt/images/${NAME}_$(date +%y%m%d).qcow2
