# pritunl-cloud: virt-install scripts

Scripts used to build base images for pritunl-cloud

```bash
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

sudo dnf -y install qemu-kvm qemu-img libguestfs-tools xorriso edk2-ovmf libvirt virt-install
sudo systemctl enable --now libvirtd

cd ./setup
sudo firewall-cmd --zone=libvirt --add-port=8000/tcp --permanent
python3 -m http.server

# alpine linux
setup-alpine
curl -o /root/setup.sh http://192.168.122.1:8000/alpine.sh
echo "c502a8b650d2b60f61414ea2f286577732ab7fc96bac487ebf024cd2120244ca /root/setup.sh" | sha256sum -c && sudo sh /root/setup.sh

# arch linux
mkdir /mnt/config
mount /dev/sr1 /mnt/config
cp /mnt/config/archinstall.json /root
umount /mnt/config
rmdir /mnt/config
pacman-key --init
pacman-key --populate archlinux
pacman -Sy --noconfirm archinstall
archinstall --silent --config /root/archinstall.json
reboot
curl -o /root/setup.sh http://192.168.122.1:8000/arch.sh
echo "412aacb35f882d09ad7390124f2e3f52a7ae8deb6aaf2825a8775912dfb058fd /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# debian
sudo curl -o /root/setup.sh http://192.168.122.1:8000/debian.sh
echo "f477fade6fb40d1767314c446ba4229111a2aba2d3db9eaea7ba86e5a8c18737 /root/setup.sh" | sudo sha256sum -c && sudo bash /root/setup.sh

# fedora
curl -o /root/setup.sh http://192.168.122.1:8000/fedora.sh
echo "bbaa3736050881897e195231ec048e2ca7b69cfa383a259c671ae3dd476fe38d /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# freebsd
fetch -o /root/setup.sh http://192.168.122.1:8000/freebsd.sh
[ "$(sha256sum /root/setup.sh)" = "aa430641df6bffd455c06258a10ce0757775ad4ab722bc4ceae3b9a773ca0827  /root/setup.sh" ] && sh /root/setup.sh

# rhel7
curl -o /root/setup.sh http://192.168.122.1:8000/rhel7.sh
echo "da5f9518e45a71f1348b7fffd14e496a64cf2bb4a73fc763ec8e97d8f4c2e6d6 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# rhel8
curl -o /root/setup.sh http://192.168.122.1:8000/rhel8.sh
echo "dd277240c6d5b573f34e98241c67caec8c0b3c855d13fbb8ccfdfda7f7e726fa /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# rhel9
curl -o /root/setup.sh http://192.168.122.1:8000/rhel9.sh
echo "23e0b0191270db7e09ade9afce206ac6a455aa7e91bf9eda6b6c677dfb78d994 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# rhel10
curl -o /root/setup.sh http://192.168.122.1:8000/rhel10.sh
echo "49cd8fd80e0a3badfbfa1b62de270e62e8f33e831e7f4780fa2f38bd89b9ffe5 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

find /var/lib/virt/images/ -name "*_$(date +%y%m%d).qcow2" -type f -exec sudo GPG_TTY=$(tty) gpg --default-key 055C08A4 --armor --output {}.sig --detach-sig {} \;

sudo mkdir -p /mnt/images
sudo chown cloud:cloud /mnt/images
mkdir -p /mnt/images/stable
mkdir -p /mnt/images/unstable
rsync --human-readable --archive --xattrs --progress 127.0.0.1:/var/lib/virt/images/ /mnt/images/unstable/

sudo wget -P /tmp https://raw.githubusercontent.com/pritunl/toolbox/73aacb9e22b09a34f87d389b3dc301d6c450b0e8/s3c/s3c.py
echo "7d14fa361e47ff328bbadac302a06a995f6ab65abbe4efce7d8cde6657ba8dde  /tmp/s3c.py" | sha256sum -c - && sudo cp /tmp/s3c.py /usr/local/bin/s3c && sudo chmod +x /usr/local/bin/s3c
sudo rm /tmp/s3c.py

cd /mnt/images/unstable
python3 ~/git/pritunl-cloud/tools/generate_files.py
python3 ~/git/pritunl-cloud/tools/autoindex.py

s3c cp almalinux8_$(date +%y%m%d).qcow2 pritunl-images:/unstable/almalinux8_$(date +%y%m%d).qcow2
s3c cp almalinux8_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/almalinux8_$(date +%y%m%d).qcow2.sig
s3c cp almalinux9_$(date +%y%m%d).qcow2 pritunl-images:/unstable/almalinux9_$(date +%y%m%d).qcow2
s3c cp almalinux9_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/almalinux9_$(date +%y%m%d).qcow2.sig
s3c cp almalinux10_$(date +%y%m%d).qcow2 pritunl-images:/unstable/almalinux10_$(date +%y%m%d).qcow2
s3c cp almalinux10_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/almalinux10_$(date +%y%m%d).qcow2.sig
s3c cp alpinelinux_$(date +%y%m%d).qcow2 pritunl-images:/unstable/alpinelinux_$(date +%y%m%d).qcow2
s3c cp alpinelinux_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/alpinelinux_$(date +%y%m%d).qcow2.sig
s3c cp archlinux_$(date +%y%m%d).qcow2 pritunl-images:/unstable/archlinux_$(date +%y%m%d).qcow2
s3c cp archlinux_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/archlinux_$(date +%y%m%d).qcow2.sig
s3c cp fedora42_$(date +%y%m%d).qcow2 pritunl-images:/unstable/fedora42_$(date +%y%m%d).qcow2
s3c cp fedora42_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/fedora42_$(date +%y%m%d).qcow2.sig
s3c cp fedora43_$(date +%y%m%d).qcow2 pritunl-images:/unstable/fedora43_$(date +%y%m%d).qcow2
s3c cp fedora43_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/fedora43_$(date +%y%m%d).qcow2.sig
s3c cp freebsd_$(date +%y%m%d).qcow2 pritunl-images:/unstable/freebsd_$(date +%y%m%d).qcow2
s3c cp freebsd_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/freebsd_$(date +%y%m%d).qcow2.sig
s3c cp oraclelinux7_$(date +%y%m%d).qcow2 pritunl-images:/unstable/oraclelinux7_$(date +%y%m%d).qcow2
s3c cp oraclelinux7_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/oraclelinux7_$(date +%y%m%d).qcow2.sig
s3c cp oraclelinux8_$(date +%y%m%d).qcow2 pritunl-images:/unstable/oraclelinux8_$(date +%y%m%d).qcow2
s3c cp oraclelinux8_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/oraclelinux8_$(date +%y%m%d).qcow2.sig
s3c cp oraclelinux9_$(date +%y%m%d).qcow2 pritunl-images:/unstable/oraclelinux9_$(date +%y%m%d).qcow2
s3c cp oraclelinux9_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/oraclelinux9_$(date +%y%m%d).qcow2.sig
s3c cp oraclelinux10_$(date +%y%m%d).qcow2 pritunl-images:/unstable/oraclelinux10_$(date +%y%m%d).qcow2
s3c cp oraclelinux10_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/oraclelinux10_$(date +%y%m%d).qcow2.sig
s3c cp rockylinux8_$(date +%y%m%d).qcow2 pritunl-images:/unstable/rockylinux8_$(date +%y%m%d).qcow2
s3c cp rockylinux8_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/rockylinux8_$(date +%y%m%d).qcow2.sig
s3c cp rockylinux9_$(date +%y%m%d).qcow2 pritunl-images:/unstable/rockylinux9_$(date +%y%m%d).qcow2
s3c cp rockylinux9_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/rockylinux9_$(date +%y%m%d).qcow2.sig
s3c cp rockylinux10_$(date +%y%m%d).qcow2 pritunl-images:/unstable/rockylinux10_$(date +%y%m%d).qcow2
s3c cp rockylinux10_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/rockylinux10_$(date +%y%m%d).qcow2.sig
s3c cp ubuntu2404_$(date +%y%m%d).qcow2 pritunl-images:/unstable/ubuntu2404_$(date +%y%m%d).qcow2
s3c cp ubuntu2404_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/ubuntu2404_$(date +%y%m%d).qcow2.sig
s3c cp files.json pritunl-images:/unstable/files.json
s3c cp index.html pritunl-images:/unstable/index.html

rsync --human-readable --archive --progress --delete /mnt/images/unstable/ /mnt/images/stable/
cd /mnt/images/stable
python3 ~/git/pritunl-cloud/tools/generate_files.py
python3 ~/git/pritunl-cloud/tools/autoindex.py

s3c cp almalinux8_$(date +%y%m%d).qcow2 pritunl-images:/stable/almalinux8_$(date +%y%m%d).qcow2
s3c cp almalinux8_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/almalinux8_$(date +%y%m%d).qcow2.sig
s3c cp almalinux9_$(date +%y%m%d).qcow2 pritunl-images:/stable/almalinux9_$(date +%y%m%d).qcow2
s3c cp almalinux9_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/almalinux9_$(date +%y%m%d).qcow2.sig
s3c cp almalinux10_$(date +%y%m%d).qcow2 pritunl-images:/stable/almalinux10_$(date +%y%m%d).qcow2
s3c cp almalinux10_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/almalinux10_$(date +%y%m%d).qcow2.sig
s3c cp alpinelinux_$(date +%y%m%d).qcow2 pritunl-images:/stable/alpinelinux_$(date +%y%m%d).qcow2
s3c cp alpinelinux_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/alpinelinux_$(date +%y%m%d).qcow2.sig
s3c cp archlinux_$(date +%y%m%d).qcow2 pritunl-images:/stable/archlinux_$(date +%y%m%d).qcow2
s3c cp archlinux_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/archlinux_$(date +%y%m%d).qcow2.sig
s3c cp fedora42_$(date +%y%m%d).qcow2 pritunl-images:/stable/fedora42_$(date +%y%m%d).qcow2
s3c cp fedora42_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/fedora42_$(date +%y%m%d).qcow2.sig
s3c cp fedora43_$(date +%y%m%d).qcow2 pritunl-images:/stable/fedora43_$(date +%y%m%d).qcow2
s3c cp fedora43_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/fedora43_$(date +%y%m%d).qcow2.sig
s3c cp freebsd_$(date +%y%m%d).qcow2 pritunl-images:/stable/freebsd_$(date +%y%m%d).qcow2
s3c cp freebsd_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/freebsd_$(date +%y%m%d).qcow2.sig
s3c cp oraclelinux7_$(date +%y%m%d).qcow2 pritunl-images:/stable/oraclelinux7_$(date +%y%m%d).qcow2
s3c cp oraclelinux7_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/oraclelinux7_$(date +%y%m%d).qcow2.sig
s3c cp oraclelinux8_$(date +%y%m%d).qcow2 pritunl-images:/stable/oraclelinux8_$(date +%y%m%d).qcow2
s3c cp oraclelinux8_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/oraclelinux8_$(date +%y%m%d).qcow2.sig
s3c cp oraclelinux9_$(date +%y%m%d).qcow2 pritunl-images:/stable/oraclelinux9_$(date +%y%m%d).qcow2
s3c cp oraclelinux9_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/oraclelinux9_$(date +%y%m%d).qcow2.sig
s3c cp oraclelinux10_$(date +%y%m%d).qcow2 pritunl-images:/stable/oraclelinux10_$(date +%y%m%d).qcow2
s3c cp oraclelinux10_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/oraclelinux10_$(date +%y%m%d).qcow2.sig
s3c cp rockylinux8_$(date +%y%m%d).qcow2 pritunl-images:/stable/rockylinux8_$(date +%y%m%d).qcow2
s3c cp rockylinux8_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/rockylinux8_$(date +%y%m%d).qcow2.sig
s3c cp rockylinux9_$(date +%y%m%d).qcow2 pritunl-images:/stable/rockylinux9_$(date +%y%m%d).qcow2
s3c cp rockylinux9_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/rockylinux9_$(date +%y%m%d).qcow2.sig
s3c cp rockylinux10_$(date +%y%m%d).qcow2 pritunl-images:/stable/rockylinux10_$(date +%y%m%d).qcow2
s3c cp rockylinux10_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/rockylinux10_$(date +%y%m%d).qcow2.sig
s3c cp ubuntu2404_$(date +%y%m%d).qcow2 pritunl-images:/stable/ubuntu2404_$(date +%y%m%d).qcow2
s3c cp ubuntu2404_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/ubuntu2404_$(date +%y%m%d).qcow2.sig
s3c cp files.json pritunl-images:/stable/files.json
s3c cp index.html pritunl-images:/stable/index.html
```
