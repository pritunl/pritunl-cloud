package qemu

const systemdTemplate = `# PritunlData=%s

[Unit]
Description=Pritunl Cloud Virtual Machine
After=network.target

[Service]
Type=simple
User=root
ExecStart=%s
`

const (
	Libvirt = "/usr/libexec/qemu-kvm"
	System  = "/usr/bin/qemu-system-x86_64"
)
