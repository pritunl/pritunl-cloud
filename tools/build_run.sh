#!/bin/bash
set -e

build_pritunl_agent() {
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -v -o agent-static
    sudo cp -f ./agent-static /usr/bin/pritunl-cloud-agent

    CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -v -o agent-bsd
    sudo cp -f ./agent-bsd /usr/bin/pritunl-cloud-agent-bsd
    rm ./agent-bsd
}

cd agent
output=$(go install -v 2>&1)
if [ -n "$output" ]; then
    build_pritunl_agent
fi

cd ..
cd redirect
go install -v
sudo cp -f ~/go/bin/redirect /usr/bin/pritunl-cloud-redirect

cd ..
cd imds/server
go install -v
sudo cp -f ~/go/bin/server /usr/bin/pritunl-cloud-imds
cd ../../

go install -v
sudo cp -f ~/go/bin/pritunl-cloud /usr/bin/pritunl-cloud

if [ $# -eq 0 ]; then
    sudo /usr/bin/pritunl-cloud start --debug
else
    sudo /usr/bin/pritunl-cloud $@
fi
