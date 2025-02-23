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

sudo sed -i 's/^SELINUX=.*/SELINUX=disabled/g' /etc/selinux/config
sudo sed -i 's/^SELINUX=.*/SELINUX=disabled/g' /etc/sysconfig/selinux
sudo setenforce 0

sudo dnf -y install qemu-kvm qemu-img libguestfs-tools genisoimage edk2-ovmf virt-install

sudo curl -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/debian.sh
echo "72ad46a18da1a52776a423cf42c0bf644faed80c32bf6a50f1c3bcfac8a992c4 /root/setup.sh" | sha256sum -c && sudo bash /root/setup.sh

curl -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/fedora.sh
echo "0cc31951dd1d1ae8530842e410a81359b8eed19db8984b0dca18c86286d685d4 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

fetch -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/freebsd.sh
[ "$(sha256sum /root/setup.sh)" = "6ed364390594913597d563330057a63baf894474c9b0649cf3b62ebb480d30e1  /root/setup.sh" ] && sh /root/setup.sh

curl -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/rhel7.sh
echo "168073816704bd6243e018a1692281b7bb6e3d10aa671d65124813539e18c4cc /root/setup.sh" | sha256sum -c && bash /root/setup.sh

curl -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/rhel8.sh
echo "5b4428d0d16fefa1857f88ae737d4a775f39b6732df1188ad180df14f29c40a6 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

curl -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/rhel9.sh
echo "4f2f2556618aaffc53261a404cc8220cbe55accf3a10da40213df3fb315f1755 /root/setup.sh" | sha256sum -c && bash /root/setup.sh

mkdir -p ~/Shared/images
scp $BUILD_IP:/var/lib/virt/images/* ~/Shared/images/

find ~/Shared/images/ -name "*.qcow2" -type f -exec gpg --default-key 055C08A4 --armor --output {}.sig --detach-sig {} \;
sha256sum ~/Shared/images/*
```
