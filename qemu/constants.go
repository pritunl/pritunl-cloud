package qemu

const systemdTemplate = `[Unit]
Description=Pritunl Cloud Virtual Machine
After=network.target

[Pritunl]
PritunlData=%s

[Service]
Type=simple
User=root
ExecStart=%s
`
