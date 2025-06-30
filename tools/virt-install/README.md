# pritunl-cloud: virt-install scripts

Scripts used to build base images for pritunl-cloud

```bash
sudo /usr/libexec/oci-growfs

sudo nano /etc/fstab
sudo umount /var/oled
sudo lvremove /dev/ocivolume/oled
sudo lvextend -l +100%FREE /dev/ocivolume/root
sudo xfs_growfs /dev/ocivolume/root

sudo tee /etc/security/limits.conf << EOF
* soft memlock 2048000000
* hard memlock 2048000000
root soft memlock 2048000000
root hard memlock 2048000000
* hard nofile 500000
* soft nofile 500000
root hard nofile 500000
root soft nofile 500000
EOF

sudo tee /etc/systemd/system/disable-thp.service << EOF
[Unit]
Description=Disable Transparent Huge Pages

[Service]
Type=simple
ExecStart=/bin/sh -c "echo 'never' > /sys/kernel/mm/transparent_hugepage/enabled && echo 'never' > /sys/kernel/mm/transparent_hugepage/defrag"

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl start disable-thp
sudo systemctl enable disable-thp

sudo sed -i 's/^SELINUX=.*/SELINUX=permissive/g' /etc/selinux/config
sudo setenforce 0

sudo dnf -y install qemu-kvm qemu-img libguestfs-tools genisoimage edk2-ovmf libvirt virt-install
sudo systemctl enable --now libvirtd

# alpine linux
setup-alpine
curl -o /root/setup.sh http://192.168.122.1:8000/alpine.sh
echo "923052cc9fd8baf7ff7eb6372391b9cb93667e70c5ddd80a6d7b0076284e3314 /root/setup.sh" | sha256sum -c && sudo sh /root/setup.sh
grub-install --target=x86_64-efi --efi-directory=/boot/efi --bootloader-id=alpine
grub-mkconfig -o /boot/grub/grub.cfg
grub-mkconfig -o /boot/efi/EFI/alpine/grub.cfg

# debian
sudo curl -o /root/setup.sh http://192.168.122.1:8000/debian.sh
echo "a4d4c35ec9aa0057373c826830944c7ec1081601f09a9af8ba432b80240210c9 /root/setup.sh" | sudo sha256sum -c && sudo bash /root/setup.sh

# fedora
curl -o /root/setup.sh http://192.168.122.1:8000/fedora.sh
echo "5577c87991735f802278ad98db35525fff820a7ff5be4b422b29eee27e362e82 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# freebsd
fetch -o /root/setup.sh http://192.168.122.1:8000/freebsd.sh
[ "$(sha256sum /root/setup.sh)" = "cd19d514fa5df15b4b7b8b66a0ced4b2ea3b3feeb710baa8cacf7c139af8ce81  /root/setup.sh" ] && sh /root/setup.sh

# rhel7
curl -o /root/setup.sh http://192.168.122.1:8000/rhel7.sh
echo "b8dc0d838f7f1a25ce1bbc9f7f326bbae1cf775d150f7a4e6bf1cc5014ada614 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# rhel8
curl -o /root/setup.sh http://192.168.122.1:8000/rhel8.sh
echo "2ebd5d85cdd4541b9901c59e1b25b270f82c88b03e9cc7d0990037549b15a27d /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# rhel9
curl -o /root/setup.sh http://192.168.122.1:8000/rhel9.sh
echo "333507a276497b69da10def3652ce3d44dd4f612578699f7a9ee1b1376855ee9 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# rhel10
curl -o /root/setup.sh http://192.168.122.1:8000/rhel10.sh
echo "13e8912e4cc96b843c49838adf26397cd8c9e2d8fe1e68793ca3fd8327b0f210 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

sudo mkdir -p /mnt/images
sudo chown cloud:cloud /mnt/images
mkdir -p /mnt/images/stable
mkdir -p /mnt/images/unstable

scp 127.0.01:/var/lib/virt/images/* /mnt/images/unstable
find /mnt/images/unstable/ -name "*.qcow2" -type f -exec gpg --default-key 055C08A4 --armor --output {}.sig --detach-sig {} \;
sha256sum /mnt/images/unstable/*

sudo wget -P /tmp https://raw.githubusercontent.com/pritunl/toolbox/73aacb9e22b09a34f87d389b3dc301d6c450b0e8/s3c/s3c.py
echo "7d14fa361e47ff328bbadac302a06a995f6ab65abbe4efce7d8cde6657ba8dde  /tmp/s3c.py" | sha256sum -c - && sudo cp /tmp/s3c.py /usr/local/bin/s3c && sudo chmod +x /usr/local/bin/s3c
sudo rm /tmp/s3c.py

cd /mnt/images/unstable
python3 ~/tools/generate_files.py
python3 ~/tools/autoindex.py

s3c almalinux8_$(date +%y%m%d).qcow2 pritunl-images/unstable/almalinux8_$(date +%y%m%d).qcow2
s3c almalinux8_$(date +%y%m%d).qcow2.sig pritunl-images/unstable/almalinux8_$(date +%y%m%d).qcow2.sig
s3c almalinux9_$(date +%y%m%d).qcow2 pritunl-images/unstable/almalinux9_$(date +%y%m%d).qcow2
s3c almalinux9_$(date +%y%m%d).qcow2.sig pritunl-images/unstable/almalinux9_$(date +%y%m%d).qcow2.sig
s3c almalinux10_$(date +%y%m%d).qcow2 pritunl-images/unstable/almalinux10_$(date +%y%m%d).qcow2
s3c almalinux10_$(date +%y%m%d).qcow2.sig pritunl-images/unstable/almalinux10_$(date +%y%m%d).qcow2.sig
s3c alpinelinux_$(date +%y%m%d).qcow2 pritunl-images/unstable/alpinelinux_$(date +%y%m%d).qcow2
s3c alpinelinux_$(date +%y%m%d).qcow2.sig pritunl-images/unstable/alpinelinux_$(date +%y%m%d).qcow2.sig
s3c fedora42_$(date +%y%m%d).qcow2 pritunl-images/unstable/fedora42_$(date +%y%m%d).qcow2
s3c fedora42_$(date +%y%m%d).qcow2.sig pritunl-images/unstable/fedora42_$(date +%y%m%d).qcow2.sig
s3c freebsd_$(date +%y%m%d).qcow2 pritunl-images/unstable/freebsd_$(date +%y%m%d).qcow2
s3c freebsd_$(date +%y%m%d).qcow2.sig pritunl-images/unstable/freebsd_$(date +%y%m%d).qcow2.sig
s3c oraclelinux7_$(date +%y%m%d).qcow2 pritunl-images/unstable/oraclelinux7_$(date +%y%m%d).qcow2
s3c oraclelinux7_$(date +%y%m%d).qcow2.sig pritunl-images/unstable/oraclelinux7_$(date +%y%m%d).qcow2.sig
s3c oraclelinux8_$(date +%y%m%d).qcow2 pritunl-images/unstable/oraclelinux8_$(date +%y%m%d).qcow2
s3c oraclelinux8_$(date +%y%m%d).qcow2.sig pritunl-images/unstable/oraclelinux8_$(date +%y%m%d).qcow2.sig
s3c oraclelinux9_$(date +%y%m%d).qcow2 pritunl-images/unstable/oraclelinux9_$(date +%y%m%d).qcow2
s3c oraclelinux9_$(date +%y%m%d).qcow2.sig pritunl-images/unstable/oraclelinux9_$(date +%y%m%d).qcow2.sig
s3c oraclelinux10_$(date +%y%m%d).qcow2 pritunl-images/unstable/oraclelinux10_$(date +%y%m%d).qcow2
s3c oraclelinux10_$(date +%y%m%d).qcow2.sig pritunl-images/unstable/oraclelinux10_$(date +%y%m%d).qcow2.sig
s3c rockylinux8_$(date +%y%m%d).qcow2 pritunl-images/unstable/rockylinux8_$(date +%y%m%d).qcow2
s3c rockylinux8_$(date +%y%m%d).qcow2.sig pritunl-images/unstable/rockylinux8_$(date +%y%m%d).qcow2.sig
s3c rockylinux9_$(date +%y%m%d).qcow2 pritunl-images/unstable/rockylinux9_$(date +%y%m%d).qcow2
s3c rockylinux9_$(date +%y%m%d).qcow2.sig pritunl-images/unstable/rockylinux9_$(date +%y%m%d).qcow2.sig
s3c rockylinux10_$(date +%y%m%d).qcow2 pritunl-images/unstable/rockylinux10_$(date +%y%m%d).qcow2
s3c rockylinux10_$(date +%y%m%d).qcow2.sig pritunl-images/unstable/rockylinux10_$(date +%y%m%d).qcow2.sig
s3c ubuntu2404_$(date +%y%m%d).qcow2 pritunl-images/unstable/ubuntu2404_$(date +%y%m%d).qcow2
s3c ubuntu2404_$(date +%y%m%d).qcow2.sig pritunl-images/unstable/ubuntu2404_$(date +%y%m%d).qcow2.sig
s3c files.json pritunl-images/unstable/files.json
s3c index.html pritunl-images/unstable/index.html

rsync --human-readable --archive --progress --delete /mnt/images/unstable/ /mnt/images/stable/
cd /mnt/images/stable
python3 ~/tools/generate_files.py
python3 ~/tools/autoindex.py

s3c almalinux8_$(date +%y%m%d).qcow2 pritunl-images/stable/almalinux8_$(date +%y%m%d).qcow2
s3c almalinux8_$(date +%y%m%d).qcow2.sig pritunl-images/stable/almalinux8_$(date +%y%m%d).qcow2.sig
s3c almalinux9_$(date +%y%m%d).qcow2 pritunl-images/stable/almalinux9_$(date +%y%m%d).qcow2
s3c almalinux9_$(date +%y%m%d).qcow2.sig pritunl-images/stable/almalinux9_$(date +%y%m%d).qcow2.sig
s3c almalinux10_$(date +%y%m%d).qcow2 pritunl-images/stable/almalinux10_$(date +%y%m%d).qcow2
s3c almalinux10_$(date +%y%m%d).qcow2.sig pritunl-images/stable/almalinux10_$(date +%y%m%d).qcow2.sig
s3c alpinelinux_$(date +%y%m%d).qcow2 pritunl-images/stable/alpinelinux_$(date +%y%m%d).qcow2
s3c alpinelinux_$(date +%y%m%d).qcow2.sig pritunl-images/stable/alpinelinux_$(date +%y%m%d).qcow2.sig
s3c fedora42_$(date +%y%m%d).qcow2 pritunl-images/stable/fedora42_$(date +%y%m%d).qcow2
s3c fedora42_$(date +%y%m%d).qcow2.sig pritunl-images/stable/fedora42_$(date +%y%m%d).qcow2.sig
s3c freebsd_$(date +%y%m%d).qcow2 pritunl-images/stable/freebsd_$(date +%y%m%d).qcow2
s3c freebsd_$(date +%y%m%d).qcow2.sig pritunl-images/stable/freebsd_$(date +%y%m%d).qcow2.sig
s3c oraclelinux7_$(date +%y%m%d).qcow2 pritunl-images/stable/oraclelinux7_$(date +%y%m%d).qcow2
s3c oraclelinux7_$(date +%y%m%d).qcow2.sig pritunl-images/stable/oraclelinux7_$(date +%y%m%d).qcow2.sig
s3c oraclelinux8_$(date +%y%m%d).qcow2 pritunl-images/stable/oraclelinux8_$(date +%y%m%d).qcow2
s3c oraclelinux8_$(date +%y%m%d).qcow2.sig pritunl-images/stable/oraclelinux8_$(date +%y%m%d).qcow2.sig
s3c oraclelinux9_$(date +%y%m%d).qcow2 pritunl-images/stable/oraclelinux9_$(date +%y%m%d).qcow2
s3c oraclelinux9_$(date +%y%m%d).qcow2.sig pritunl-images/stable/oraclelinux9_$(date +%y%m%d).qcow2.sig
s3c oraclelinux10_$(date +%y%m%d).qcow2 pritunl-images/stable/oraclelinux10_$(date +%y%m%d).qcow2
s3c oraclelinux10_$(date +%y%m%d).qcow2.sig pritunl-images/stable/oraclelinux10_$(date +%y%m%d).qcow2.sig
s3c rockylinux8_$(date +%y%m%d).qcow2 pritunl-images/stable/rockylinux8_$(date +%y%m%d).qcow2
s3c rockylinux8_$(date +%y%m%d).qcow2.sig pritunl-images/stable/rockylinux8_$(date +%y%m%d).qcow2.sig
s3c rockylinux9_$(date +%y%m%d).qcow2 pritunl-images/stable/rockylinux9_$(date +%y%m%d).qcow2
s3c rockylinux9_$(date +%y%m%d).qcow2.sig pritunl-images/stable/rockylinux9_$(date +%y%m%d).qcow2.sig
s3c rockylinux10_$(date +%y%m%d).qcow2 pritunl-images/stable/rockylinux10_$(date +%y%m%d).qcow2
s3c rockylinux10_$(date +%y%m%d).qcow2.sig pritunl-images/stable/rockylinux10_$(date +%y%m%d).qcow2.sig
s3c ubuntu2404_$(date +%y%m%d).qcow2 pritunl-images/stable/ubuntu2404_$(date +%y%m%d).qcow2
s3c ubuntu2404_$(date +%y%m%d).qcow2.sig pritunl-images/stable/ubuntu2404_$(date +%y%m%d).qcow2.sig
s3c files.json pritunl-images/stable/files.json
s3c index.html pritunl-images/stable/index.html
```
