#!/bin/bash
set -ev
NAME="archlinux"
ISO_URL="https://geo.mirror.pkgbuild.com/iso/latest/archlinux-2025.12.01-x86_64.iso"
ISO_HASH="c2b1f13a68482db3aad008f14bb75cb15a44cd38fa8a1aa15e6675a50d4c4374"

sudo mkdir -p /var/lib/virt/iso
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

sudo tee /var/lib/virt/init/${NAME}/archinstall.json << EOF
{
    "additional-repositories": null,
    "archinstall-language": "English",
    "audio_config": null,
    "bootloader": "Systemd-boot",
    "debug": false,
    "disk_config": {
        "config_type": "manual_partitioning",
        "device_modifications": [
            {
                "device": "/dev/vda",
                "partitions": [
                    {
                        "btrfs": [],
                        "dev_path": null,
                        "flags": ["boot", "esp"],
                        "fs_type": "fat32",
                        "size": {
                            "sector_size": {
                                "unit": "B",
                                "value": 512
                            },
                            "unit": "MiB",
                            "value": 512
                        },
                        "mount_options": [],
                        "mountpoint": "/boot",
                        "obj_id": "$(uuidgen)",
                        "start": {
                            "sector_size": {
                                "unit": "B",
                                "value": 512
                            },
                            "unit": "MiB",
                            "value": 1
                        },
                        "status": "create",
                        "type": "primary"
                    },
                    {
                        "btrfs": [],
                        "dev_path": null,
                        "flags": [],
                        "fs_type": "xfs",
                        "size": {
                            "sector_size": {
                                "unit": "B",
                                "value": 512
                            },
                            "unit": "MiB",
                            "value": 7678
                        },
                        "mount_options": [],
                        "mountpoint": "/",
                        "obj_id": "$(uuidgen)",
                        "start": {
                            "sector_size": {
                                "unit": "B",
                                "value": 512
                            },
                            "unit": "MiB",
                            "value": 513
                        },
                        "status": "create",
                        "type": "primary"
                    }
                ],
                "wipe": true
            }
        ]
    },
    "hostname": "cloud",
    "kernels": ["linux"],
    "locale_config": {
        "kb_layout": "us",
        "sys_enc": "UTF-8",
        "sys_lang": "en_US"
    },
    "network_config": {
        "type": "nm"
    },
    "no_pkg_lookups": false,
    "ntp": true,
    "offline": false,
    "packages": [
        "base",
        "base-devel",
        "linux",
        "linux-firmware",
        "networkmanager",
        "openssh",
        "efibootmgr",
        "vi"
    ],
    "parallel downloads": 0,
    "profile_config": {
        "gfx_driver": null,
        "greeter": null,
        "profile": {
            "custom_settings": {},
            "details": [],
            "main": "Server"
        }
    },
    "mirror_config": {
        "custom_mirrors": [],
        "mirror_regions": {
            "Worldwide": [
                "https://geo.mirror.pkgbuild.com/\$repo/os/\$arch",
                "https://mirror.rackspace.com/archlinux/\$repo/os/\$arch"
            ]
        }
    },
    "custom_commands": [
        "sed -i 's/rootfstype=xfs\$/rootfstype=xfs console=ttyS0/' /boot/loader/entries/*.conf",
        "systemctl enable serial-getty@ttyS0.service",
        "systemctl set-default multi-user.target",
        "mkinitcpio -P"
    ],
    "save_config": null,
    "script": "guided",
    "silent": true,
    "swap": false,
    "timezone": "UTC",
    "version": "2.8.6",
    "root_enc_password": "\$y\$j9T\$KsJ3WRqoGvcjGsQNis/oG0\$0zE1DqJ4NJn6pEN3VhnaUIA/nIBSeIYNR8yShbphLW1"
}
EOF

sudo rm -f /var/lib/virt/init/${NAME}/archinstall.iso
sudo xorriso -as mkisofs \
  -output /var/lib/virt/init/${NAME}/archinstall.iso \
  -volid CONFIGDATA \
  -joliet \
  -rock \
  -input-charset utf-8 \
  /var/lib/virt/init/${NAME}/archinstall.json

sudo tee /var/lib/virt/init/${NAME}/archinstall-auto.service << 'EOF'
[Unit]
Description=Automated Arch Installation
After=multi-user.target

[Service]
Type=oneshot
ExecStart=/usr/bin/archinstall --silent --config /archinstall-config.json
ExecStartPost=/usr/bin/systemctl poweroff

[Install]
WantedBy=multi-user.target
EOF

sudo tee /var/lib/virt/init/${NAME}/archinstall-auto.service << 'EOF'
[Unit]
Description=Automated Arch Installation
After=multi-user.target

[Service]
Type=oneshot
ExecStart=/usr/bin/archinstall --config /archinstall-config.json --silent
ExecStartPost=/usr/bin/systemctl poweroff

[Install]
WantedBy=multi-user.target
EOF

sudo tee /var/lib/virt/init/${NAME}/autoinstall.sh << 'EOF'
#!/bin/bash
set -ex

mkdir /mnt/config
mount /dev/sr1 /mnt/config
cp /mnt/config/archinstall.json /root/
umount /mnt/config
rmdir /mnt/config

archinstall --silent --config /root/archinstall.json

reboot
EOF
sudo chmod +x /var/lib/virt/init/${NAME}/autoinstall.sh

sudo virt-install \
  --name ${NAME} \
  --vcpus 8 \
  --memory 8192 \
  --boot uefi,firmware.feature0.name=secure-boot,firmware.feature0.enabled=no \
  --disk path=/var/lib/virt/${NAME}.qcow2,size=8,format=qcow2,bus=virtio \
  --disk path=/var/lib/virt/init/${NAME}/archinstall.iso,device=cdrom,bus=sata \
  --os-variant archlinux \
  --network network=default \
  --graphics=none \
  --console pty,target_type=serial \
  --location=/var/lib/virt/iso/$(basename ${ISO_URL}),kernel=arch/boot/x86_64/vmlinuz-linux,initrd=arch/boot/x86_64/initramfs-linux.img \
  --extra-args="console=ttyS0 archisobasedir=arch archisosearchuuid=$(blkid -s UUID -o value /var/lib/virt/iso/$(basename ${ISO_URL})) cow_spacesize=1G"

while ! sudo virsh domstate ${NAME} 2>/dev/null | grep -q "shut off"; do
  sleep 1
done

sudo rm -rf /var/lib/virt/init/${NAME}

echo "Compressing image..."
sudo rm -f /var/lib/virt/images/${NAME}_$(date +%y%m%d).qcow2
sudo qemu-img convert -f qcow2 -O qcow2 -c /var/lib/virt/${NAME}.qcow2 /var/lib/virt/images/${NAME}_$(date +%y%m%d).qcow2
sha256sum /var/lib/virt/images/${NAME}_$(date +%y%m%d).qcow2
