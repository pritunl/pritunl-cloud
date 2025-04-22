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

sudo dnf -y install qemu-kvm qemu-img libguestfs-tools genisoimage edk2-ovmf virt-install

# alpine linux
setup-alpine
curl -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/alpine.sh
echo "45b5e4f7821a34bbb858a664066e5c71f6bc586903ca3dc7f5e6429a3b4ba84f /root/setup.sh" | sha256sum -c && sudo sh /root/setup.sh

# debian
sudo curl -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/debian.sh
echo "181e0a2c0faddc79d8e2565782b58c0e94683904fa0dd41f63f8b6ddfcea6ab0 /root/setup.sh" | sudo sha256sum -c && sudo bash /root/setup.sh

# fedora
curl -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/fedora.sh
echo "96a21a8a4d46225263bf2679a7aacf0feeefd42c87e2c87d3dc5b3ff53833b06 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# freebsd
fetch -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/freebsd.sh
[ "$(sha256sum /root/setup.sh)" = "f6f9445e3cd6eb04575593d0c98136251f9860c1ea684f63f927bb52f5f0c702  /root/setup.sh" ] && sh /root/setup.sh

# rhel7
curl -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/rhel7.sh
echo "7f89af3f4553ae500bec9fa154bf86d2c22e4ef79fdfd514841c24dcbdcc95d3 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# rhel8
curl -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/rhel8.sh
echo "422045cc0beb1756489b67f237325b44d05db869b3b2e77f68db65e5cf7566e1 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# rhel9
curl -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/rhel9.sh
echo "38301f86bb05072ee041b53004266d38622a6503d2438b6ecb2865771c9fda84 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

sudo mkdir -p /mnt/images
sudo chown cloud:cloud /mnt/images
mkdir -p /mnt/images/stable
mkdir -p /mnt/images/unstable

scp 127.0.01:/var/lib/virt/images/* /mnt/images/unstable
find /mnt/images/unstable/ -name "*.qcow2" -type f -exec gpg --default-key 055C08A4 --armor --output {}.sig --detach-sig {} \;
sha256sum /mnt/images/unstable/*

cd /mnt/images/unstable
python3 ~/git/pritunl-cloud/tools/generate_files.py
python3 ~/git/pritunl-cloud/tools/autoindex.py

python3 ~/git/pritunl-cloud/tools/s3_upload.py almalinux8_2504.qcow2 pritunl-images/unstable/almalinux8_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py almalinux8_2504.qcow2.sig pritunl-images/unstable/almalinux8_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py almalinux9_2504.qcow2 pritunl-images/unstable/almalinux9_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py almalinux9_2504.qcow2.sig pritunl-images/unstable/almalinux9_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py alpinelinux_2504.qcow2 pritunl-images/unstable/alpinelinux_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py alpinelinux_2504.qcow2.sig pritunl-images/unstable/alpinelinux_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py fedora42_2504.qcow2 pritunl-images/unstable/fedora42_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py fedora42_2504.qcow2.sig pritunl-images/unstable/fedora42_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py freebsd_2504.qcow2 pritunl-images/unstable/freebsd_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py freebsd_2504.qcow2.sig pritunl-images/unstable/freebsd_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py oraclelinux7_2504.qcow2 pritunl-images/unstable/oraclelinux7_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py oraclelinux7_2504.qcow2.sig pritunl-images/unstable/oraclelinux7_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py oraclelinux8_2504.qcow2 pritunl-images/unstable/oraclelinux8_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py oraclelinux8_2504.qcow2.sig pritunl-images/unstable/oraclelinux8_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py oraclelinux9_2504.qcow2 pritunl-images/unstable/oraclelinux9_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py oraclelinux9_2504.qcow2.sig pritunl-images/unstable/oraclelinux9_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py rockylinux8_2504.qcow2 pritunl-images/unstable/rockylinux8_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py rockylinux8_2504.qcow2.sig pritunl-images/unstable/rockylinux8_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py rockylinux9_2504.qcow2 pritunl-images/unstable/rockylinux9_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py rockylinux9_2504.qcow2.sig pritunl-images/unstable/rockylinux9_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py ubuntu2404_2504.qcow2 pritunl-images/unstable/ubuntu2404_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py ubuntu2404_2504.qcow2.sig pritunl-images/unstable/ubuntu2404_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py files.json pritunl-images/unstable/files.json
python3 ~/git/pritunl-cloud/tools/s3_upload.py index.html pritunl-images/unstable/index.html

rsync --human-readable --archive --progress --delete /mnt/images/unstable/ /mnt/images/stable/
cd /mnt/images/stable
python3 ~/git/pritunl-cloud/tools/generate_files.py
python3 ~/git/pritunl-cloud/tools/autoindex.py

python3 ~/git/pritunl-cloud/tools/s3_upload.py almalinux8_2504.qcow2 pritunl-images/stable/almalinux8_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py almalinux8_2504.qcow2.sig pritunl-images/stable/almalinux8_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py almalinux9_2504.qcow2 pritunl-images/stable/almalinux9_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py almalinux9_2504.qcow2.sig pritunl-images/stable/almalinux9_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py alpinelinux_2504.qcow2 pritunl-images/stable/alpinelinux_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py alpinelinux_2504.qcow2.sig pritunl-images/stable/alpinelinux_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py fedora42_2504.qcow2 pritunl-images/stable/fedora42_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py fedora42_2504.qcow2.sig pritunl-images/stable/fedora42_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py freebsd_2504.qcow2 pritunl-images/stable/freebsd_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py freebsd_2504.qcow2.sig pritunl-images/stable/freebsd_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py oraclelinux7_2504.qcow2 pritunl-images/stable/oraclelinux7_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py oraclelinux7_2504.qcow2.sig pritunl-images/stable/oraclelinux7_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py oraclelinux8_2504.qcow2 pritunl-images/stable/oraclelinux8_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py oraclelinux8_2504.qcow2.sig pritunl-images/stable/oraclelinux8_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py oraclelinux9_2504.qcow2 pritunl-images/stable/oraclelinux9_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py oraclelinux9_2504.qcow2.sig pritunl-images/stable/oraclelinux9_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py rockylinux8_2504.qcow2 pritunl-images/stable/rockylinux8_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py rockylinux8_2504.qcow2.sig pritunl-images/stable/rockylinux8_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py rockylinux9_2504.qcow2 pritunl-images/stable/rockylinux9_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py rockylinux9_2504.qcow2.sig pritunl-images/stable/rockylinux9_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py ubuntu2404_2504.qcow2 pritunl-images/stable/ubuntu2404_2504.qcow2
python3 ~/git/pritunl-cloud/tools/s3_upload.py ubuntu2404_2504.qcow2.sig pritunl-images/stable/ubuntu2404_2504.qcow2.sig
python3 ~/git/pritunl-cloud/tools/s3_upload.py files.json pritunl-images/stable/files.json
python3 ~/git/pritunl-cloud/tools/s3_upload.py index.html pritunl-images/stable/index.html
```
