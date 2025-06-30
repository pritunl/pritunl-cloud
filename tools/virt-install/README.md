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

cd ./setup
sudo firewall-cmd --zone=libvirt --add-port=8000/tcp --permanent
python3 -m http.server

# alpine linux
setup-alpine
curl -o /root/setup.sh http://192.168.122.1:8000/alpine.sh
echo "f39b6192f43c6e62fafe7b9521bc1170bdf19726b71fb6dbb71b77b77dbbee13 /root/setup.sh" | sha256sum -c && sudo sh /root/setup.sh

# debian
sudo curl -o /root/setup.sh http://192.168.122.1:8000/debian.sh
echo "e950965dcdc2f7a9c415800a2b6fd07877d2b6b4f04ad74f7e5d78bedd6141c2 /root/setup.sh" | sudo sha256sum -c && sudo bash /root/setup.sh

# fedora
curl -o /root/setup.sh http://192.168.122.1:8000/fedora.sh
echo "4863da2df2fe1f14efcebdc28cbbaedf8586ab5c67695142360842e78bd732fd /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# freebsd
fetch -o /root/setup.sh http://192.168.122.1:8000/freebsd.sh
[ "$(sha256sum /root/setup.sh)" = "2efcf4a3c01618cd47b6955effc4b55600b049fef30c8a04fbf9f79250ad9b32  /root/setup.sh" ] && sh /root/setup.sh

# rhel7
curl -o /root/setup.sh http://192.168.122.1:8000/rhel7.sh
echo "396426b0db2833d81b4b12e8e7dd9aeca5ed77fc81e696860b86604fd735be31 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# rhel8
curl -o /root/setup.sh http://192.168.122.1:8000/rhel8.sh
echo "c4ed6bb54694a9e2fdc1b59806dff75eb789a8ee996b801dd5b39da3db5fddf9 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# rhel9
curl -o /root/setup.sh http://192.168.122.1:8000/rhel9.sh
echo "c482aa3906b9c168d066753855bbff4caf4e2ba3c0ce698edaf57343540b030f /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# rhel10
curl -o /root/setup.sh http://192.168.122.1:8000/rhel10.sh
echo "f97af43a199b228a0b7ed08cabfb2cb6b6718a2b089b7ebbf49d655af8430feb /root/setup.sh" | sha256sum -c && bash /root/setup.sh

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
s3c cp fedora42_$(date +%y%m%d).qcow2 pritunl-images:/unstable/fedora42_$(date +%y%m%d).qcow2
s3c cp fedora42_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/fedora42_$(date +%y%m%d).qcow2.sig
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
s3c cp fedora42_$(date +%y%m%d).qcow2 pritunl-images:/stable/fedora42_$(date +%y%m%d).qcow2
s3c cp fedora42_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/fedora42_$(date +%y%m%d).qcow2.sig
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
