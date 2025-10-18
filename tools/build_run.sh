#!/bin/bash
set -e

NO_AGENT=false
STABLE=false
ARGS=()

while [[ $# -gt 0 ]]; do
    case $1 in
        --no-agent)
            NO_AGENT=true
            shift
            ;;
        --stable)
            STABLE=true
            shift
            ;;
        *)
            ARGS+=("$1")
            shift
            ;;
    esac
done

build_pritunl_agent() {
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -v -o agent-static
    sudo cp -f ./agent-static /usr/bin/pritunl-cloud-agent
    rm -f ./agent-static

    CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -v -o agent-bsd
    sudo cp -f ./agent-bsd /usr/bin/pritunl-cloud-agent-bsd
    rm -f ./agent-bsd
}

if [ "$NO_AGENT" = false ]; then
    cd agent
    output=$(go install -v 2>&1 | tee /dev/tty) || exit 1
    if [ -n "$output" ]; then
        build_pritunl_agent
    fi
    cd ..
fi

cd redirect
go install -v
sudo cp -f ~/go/bin/redirect /usr/bin/pritunl-cloud-redirect
cd ..

go install -v
sudo cp -f ~/go/bin/pritunl-cloud /usr/bin/pritunl-cloud

if [ "$STABLE" = true ]; then
    sudo /usr/bin/pritunl-cloud start --fast-exit
elif [ ${#ARGS[@]} -eq 0 ]; then
    sudo /usr/bin/pritunl-cloud start --debug
else
    sudo /usr/bin/pritunl-cloud "${ARGS[@]}"
fi
