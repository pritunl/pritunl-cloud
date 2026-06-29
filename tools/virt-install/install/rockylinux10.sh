#!/bin/bash
set -ev
NAME="rockylinux10"
ISO_URL="https://sjc.mirror.rackspace.com/rocky/10/isos/x86_64/Rocky-10.2-x86_64-dvd1.iso"
ISO_HASH="16ca9c96cdb221ba6e1f68579f21bd69fd8da81c6a921d9068949796f91c8feb"

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

sudo tee /var/lib/virt/ks/${NAME}.ks << EOF
text
cdrom

%addon com_redhat_kdump --disable
%end

keyboard --xlayouts='us'
lang en_US.UTF-8

network --bootproto=dhcp --hostname=cloud --activate

%packages
@^minimal-environment
@standard
-kexec-tools
%end

firstboot --enable

ignoredisk --only-use=vda
clearpart --all --initlabel
part /boot/efi --fstype="efi" --ondisk=vda --size=100 --fsoptions="umask=0077,shortname=winnt"
part / --fstype="xfs" --ondisk=vda --grow

timezone Etc/UTC --utc

rootpw --plaintext cloud

%post
grubby --update-kernel=ALL --remove-args=crashkernel
%end
EOF

sudo virt-install \
  --name ${NAME} \
  --vcpus 8 \
  --memory 8192 \
  --boot uefi \
  --disk path=/var/lib/virt/${NAME}.qcow2,size=8,format=qcow2,bus=virtio \
  --os-variant rocky9 \
  --network network=default \
  --graphics=none \
  --console pty,target_type=serial \
  --location=/var/lib/virt/iso/$(basename ${ISO_URL}) \
  --initrd-inject=/var/lib/virt/ks/${NAME}.ks \
  --extra-args="console=ttyS0 inst.ks=file:/${NAME}.ks inst.text"

while ! sudo virsh domstate ${NAME} 2>/dev/null | grep -q "shut off"; do
  sleep 1
done

sudo rm -rf /var/lib/virt/init/${NAME}

echo "Compressing image..."
sudo rm -f /var/lib/virt/images/${NAME}_$(date +%y%m%d).qcow2
sudo qemu-img convert -f qcow2 -O qcow2 -c /var/lib/virt/${NAME}.qcow2 /var/lib/virt/images/${NAME}_$(date +%y%m%d).qcow2
sha256sum /var/lib/virt/images/${NAME}_$(date +%y%m%d).qcow2
