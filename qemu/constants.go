package qemu

const systemdTemplate = `# PritunlData=%s

[Unit]
Description=Pritunl Cloud Virtual Machine
After=network.target

[Service]%s
Type=simple
User=root
ExecStart=%s
PrivateTmp=%s
ProtectHome=%s
ProtectSystem=full
ProtectHostname=true
ProtectKernelTunables=true
PrivateIPC=true
NetworkNamespacePath=/var/run/netns/%s
`

const systemdTemplateExternalNet = `# PritunlData=%s

[Unit]
Description=Pritunl Cloud Virtual Machine
After=network.target

[Service]%s
Type=simple
User=root
ExecStart=%s
PrivateTmp=%s
ProtectHome=%s
ProtectSystem=full
ProtectHostname=true
ProtectKernelTunables=true
PrivateIPC=true
`
