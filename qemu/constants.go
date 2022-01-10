package qemu

const systemdTemplate = `# PritunlData=%s

[Unit]
Description=Pritunl Cloud Virtual Machine
After=network.target

[Service]%s
Type=simple
User=root
ExecStart=%s
`
