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

sh almalinux8.sh
sh almalinux9.sh
sh almalinux10.sh
sh oraclelinux7.sh
sh oraclelinux8.sh
sh oraclelinux9.sh
sh oraclelinux10.sh
sh rockylinux8.sh
sh rockylinux9.sh
sh rockylinux10.sh

sh fedora43.sh
sh fedora44.sh

sh alpinelinux.sh
sh archlinux.sh

sh ubuntu24.sh
sh ubuntu26.sh

sh freebsd.sh

# alpine linux
setup-alpine
curl -o /root/setup.sh http://192.168.122.1:8000/alpine.sh
echo "5632783ae0ecdcde28dbcb960eb69e1900d53f3197a99e83823d06b0ec90f652 /root/setup.sh" | sha256sum -c && sudo sh /root/setup.sh

# arch linux
mkdir /mnt/config
mount /dev/sr1 /mnt/config
cp /mnt/config/archinstall.json /root
umount /mnt/config
rmdir /mnt/config
archinstall --silent --config /root/archinstall.json
reboot
curl -o /root/setup.sh http://192.168.122.1:8000/arch.sh
echo "7385d5a8093b4c1363870a6454e34768aa4f8158b91e0a4bd2988b8b1a9bdaa9 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# debian
sudo curl -o /root/setup.sh http://192.168.122.1:8000/debian.sh
echo "cfd58b0b0a4c5da755f5f28eb71159fc32c75836a058ce8927cd3f3b853754ea /root/setup.sh" | sudo sha256sum -c && sudo bash /root/setup.sh

# fedora
curl -o /root/setup.sh http://192.168.122.1:8000/fedora.sh
echo "ef59ae6a30463a20798458e0e9c389b004d531a8fabd533fef5223017250691b /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# freebsd
fetch -o /root/setup.sh http://192.168.122.1:8000/freebsd.sh
[ "$(sha256sum /root/setup.sh)" = "6aab203e3ba7c8aa31ad9dc7da38f701f3871a1ae339904e1f1e6f774ec58238  /root/setup.sh" ] && sh /root/setup.sh

# rhel7
curl -o /root/setup.sh http://192.168.122.1:8000/rhel7.sh
echo "2d2b894121c99fba5556944ac18ada1a93865260539543c092721be00ea9ad62 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# rhel8
curl -o /root/setup.sh http://192.168.122.1:8000/rhel8.sh
echo "a898951fb81bbb460ea063839a077a2174af49de1cca767374027bbc3bb6afb4 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# rhel9
curl -o /root/setup.sh http://192.168.122.1:8000/rhel9.sh
echo "9c184ecc22c144904bf1ee11b13dfbb757d41a3432f80d8427177b8b8162620d /root/setup.sh" | sha256sum -c && bash /root/setup.sh

# rhel10
curl -o /root/setup.sh http://192.168.122.1:8000/rhel10.sh
echo "859e3a0bb0114c2f68096f7ec5013fe42e5578b049df8968eec1c7ad78bcc8e5 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

find /var/lib/virt/images/ -name "*_$(date +%y%m%d).qcow2" -type f -exec sudo GPG_TTY=$(tty) gpg --default-key 055C08A4 --armor --output {}.sig --detach-sig {} \;

sudo mkdir -p /mnt/images
sudo chown cloud:cloud /mnt/images
mkdir -p /mnt/images/stable
mkdir -p /mnt/images/unstable
rsync --human-readable --archive --progress image.panda.pritunl.net:/var/lib/virt/images/ /mnt/images/unstable/

sudo wget -P /tmp https://raw.githubusercontent.com/pritunl/toolbox/eae247b569672382df53d8d42ee843c219342097/s3c/s3c.py
echo "fc7680352667822e3aa3cffa7678a6f8352a340daf1b72c356aa3c95e8ba9729  /tmp/s3c.py" | sha256sum -c - && sudo cp /tmp/s3c.py /usr/local/bin/s3c && sudo chmod +x /usr/local/bin/s3c
sudo rm /tmp/s3c.py

cd /mnt/images/unstable
python3 ~/git/pritunl-cloud/tools/generate_files.py

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
s3c cp fedora43_$(date +%y%m%d).qcow2 pritunl-images:/unstable/fedora43_$(date +%y%m%d).qcow2
s3c cp fedora43_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/fedora43_$(date +%y%m%d).qcow2.sig
s3c cp fedora44_$(date +%y%m%d).qcow2 pritunl-images:/unstable/fedora44_$(date +%y%m%d).qcow2
s3c cp fedora44_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/fedora44_$(date +%y%m%d).qcow2.sig
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
s3c cp ubuntu2604_$(date +%y%m%d).qcow2 pritunl-images:/unstable/ubuntu2604_$(date +%y%m%d).qcow2
s3c cp ubuntu2604_$(date +%y%m%d).qcow2.sig pritunl-images:/unstable/ubuntu2604_$(date +%y%m%d).qcow2.sig
s3c cp files.json pritunl-images:/unstable/files.json
s3c index pritunl-images:/unstable/

rsync --human-readable --archive --progress --delete /mnt/images/unstable/ /mnt/images/stable/
cd /mnt/images/stable
python3 ~/git/pritunl-cloud/tools/generate_files.py

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
s3c cp fedora43_$(date +%y%m%d).qcow2 pritunl-images:/stable/fedora43_$(date +%y%m%d).qcow2
s3c cp fedora43_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/fedora43_$(date +%y%m%d).qcow2.sig
s3c cp fedora44_$(date +%y%m%d).qcow2 pritunl-images:/stable/fedora44_$(date +%y%m%d).qcow2
s3c cp fedora44_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/fedora44_$(date +%y%m%d).qcow2.sig
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
s3c cp ubuntu2604_$(date +%y%m%d).qcow2 pritunl-images:/stable/ubuntu2604_$(date +%y%m%d).qcow2
s3c cp ubuntu2604_$(date +%y%m%d).qcow2.sig pritunl-images:/stable/ubuntu2604_$(date +%y%m%d).qcow2.sig
s3c cp files.json pritunl-images:/stable/files.json
s3c index pritunl-images:/stable/
```
