# pritunl-cloud: declarative kvm virtualization

[![github](https://img.shields.io/badge/github-pritunl-11bdc2.svg?style=flat)](https://github.com/pritunl)
[![twitter](https://img.shields.io/badge/twitter-pritunl-55acee.svg?style=flat)](https://twitter.com/pritunl)
[![medium](https://img.shields.io/badge/medium-pritunl-b32b2b.svg?style=flat)](https://pritunl.medium.com)
[![forum](https://img.shields.io/badge/discussion-forum-ffffff.svg?style=flat)](https://forum.pritunl.com)

[Pritunl-Cloud](https://cloud.pritunl.com) is a declarative KVM virtualization
platform with shell and python based live updating templates. Documentation
and more information can be found at
[docs.pritunl.com](https://docs.pritunl.com/kb/cloud)

[![pritunl](img/logo_code.png)](https://docs.pritunl.com/kb/cloud)

## Install from Source

```bash
# Install Go
sudo dnf -y install git-core iptables net-tools ipset ipvsadm xorriso qemu-kvm qemu-img swtpm-tools

sudo rm -rf /usr/local/go
wget https://go.dev/dl/go1.25.4.linux-amd64.tar.gz
echo "9fa5ffeda4170de60f67f3aa0f824e426421ba724c21e133c1e35d6159ca1bec go1.25.4.linux-amd64.tar.gz" | sha256sum -c - && sudo tar -C /usr/local -xf go1.25.4.linux-amd64.tar.gz
rm -f go1.25.4.linux-amd64.tar.gz

tee -a ~/.bashrc << EOF
export GOPATH=\$HOME/go
export GOROOT=/usr/local/go
export PATH=/usr/local/go/bin:\$PATH
EOF
source ~/.bashrc

# Install MongoDB
sudo tee /etc/yum.repos.d/mongodb-org.repo << EOF
[mongodb-org]
name=MongoDB Repository
baseurl=https://repo.mongodb.org/yum/redhat/9/mongodb-org/8.0/x86_64/
gpgcheck=1
enabled=1
gpgkey=https://pgp.mongodb.com/server-8.0.asc
EOF

sudo dnf -y install mongodb-org
sudo systemctl enable --now mongod

# Build Pritunl Cloud
go install -v github.com/pritunl/pritunl-cloud@latest
go install -v github.com/pritunl/pritunl-cloud/redirect@latest
go install -v github.com/pritunl/pritunl-cloud/agent@latest
GOOS=freebsd GOARCH=amd64 go install -v github.com/pritunl/pritunl-cloud/agent@latest

# Setup systemd units
sudo cp $(ls -d ~/go/pkg/mod/github.com/pritunl/pritunl-cloud@v* | sort -V | tail -n 1)/tools/pritunl-cloud.service /etc/systemd/system/
sudo cp $(ls -d ~/go/pkg/mod/github.com/pritunl/pritunl-cloud@v* | sort -V | tail -n 1)/tools/pritunl-cloud-redirect.socket /etc/systemd/system/
sudo cp $(ls -d ~/go/pkg/mod/github.com/pritunl/pritunl-cloud@v* | sort -V | tail -n 1)/tools/pritunl-cloud-redirect.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo useradd -r -s /sbin/nologin -c 'Pritunl web server' pritunl-cloud-web

# Install Pritunl Cloud
sudo mkdir -p /usr/share/pritunl-cloud/www/
sudo cp -r $(ls -d ~/go/pkg/mod/github.com/pritunl/pritunl-cloud@v* | sort -V | tail -n 1)/www/dist/. /usr/share/pritunl-cloud/www/
sudo cp ~/go/bin/pritunl-cloud /usr/bin/pritunl-cloud
sudo cp ~/go/bin/redirect /usr/bin/pritunl-cloud-redirect
sudo cp ~/go/bin/agent /usr/bin/pritunl-cloud-agent
sudo cp ~/go/bin/freebsd_amd64/agent /usr/bin/pritunl-cloud-agent-bsd

sudo systemctl enable --now pritunl-cloud
```

## License

Please refer to the [`LICENSE`](LICENSE) file for a copy of the license.
