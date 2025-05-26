package qemu

const systemdTemplate = `# PritunlData=%s

[Unit]
Description=Pritunl Cloud Virtual Machine
After=network.target

[Service]%s
Environment=XDG_CACHE_HOME=%s
Type=simple
User=root
ExecStart=%s
TimeoutStopSec=5
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
Environment=XDG_CACHE_HOME=%s
Type=simple
User=root
ExecStart=%s
TimeoutStopSec=5
PrivateTmp=%s
ProtectHome=%s
ProtectSystem=full
ProtectHostname=true
ProtectKernelTunables=true
PrivateIPC=true
`
