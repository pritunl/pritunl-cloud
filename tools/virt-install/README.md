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

sudo dnf -y install qemu-kvm qemu-img libguestfs-tools genisoimage edk2-ovmf virt-install
```

```bash
curl -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/debian.sh
echo "6c5648e442a0027d2f2ac15ea3104155086af4f1df5b2cb2e59fb04cecf639ec setup.sh" | sha256sum -c && sudo bash /root/setup.sh

curl -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/rhel7.sh
echo "6553eae61cbfe23d58d6a7c462a0fd12a20e2e3e4387facbf09226519fc421e1 setup.sh" | sha256sum -c && sudo bash /root/setup.sh

curl -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/rhel8.sh
echo "2773694d6a523643a1d100d46dffd177df8e868fde6fb663af8b8a37be32d786 setup.sh" | sha256sum -c && sudo bash /root/setup.sh

curl -o /root/setup.sh https://raw.githubusercontent.com/pritunl/pritunl-cloud/refs/heads/master/tools/virt-install/setup/rhel9.sh
echo "930c1304036d3eec02d7ccad08308182a7bdf4fd17d99f770dc2b2c173661260 setup.sh" | sha256sum -c && sudo bash /root/setup.sh

mkdir -p ~/Shared/images
scp $BUILD_IP:/var/lib/virt/images/* ~/Shared/images/

find ~/Shared/images/ -name "*.qcow2" -type f -exec gpg --default-key 055C08A4 --armor --output {}.sig --detach-sig {} \;
```
