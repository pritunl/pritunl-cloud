package demo

import (
	"github.com/pritunl/pritunl-cloud/authority"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Authorities = []*authority.Authority{
	{
		Id:           utils.ObjectIdHex("688ab80d1793930f821f4f3c"),
		Name:         "cloud",
		Comment:      "",
		Type:         "ssh_key",
		Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Roles:        []string{"instance"},
		Key:          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDE+MfxSp/77B/CphRquYSxnq7ee6EbcDASJHFeMx8MVoneQ0TRDdLsUJ3KwVKvx9VcXOzqOlV6SQnT/Zwfhz8NDeqePYWeN7tI4rMVRSmlg7wYj7affVpIWXmuMSfdmZytr/PCr4h/CwjA+KJdpIBqU9B0enosTt5+OwLmViL1VLk6oi6C9UNQyszx9btOZfYnEVZ+sm7iWVaqO4Z4An7cM4V9dzbT4GOI2F1AYp+RfdvktEccCcjzZSbyJkhRt5DkRx9q/PbNwF4bRNw9gKjAcxYt54BeJ0By1HUd1snTblftlN3CQskKNXlFI7fFqQfLpaO4csi8dWu8IH7Td5YV5MKlSt+ljoNhBE+5bntQWjGU209PS/DGRW62LTIF9tiCTvJhh+fmof0mKJlABH6Es1Qzr6iwKz22a8LunF9Sf/dm9Og8zZVuCPWJNWIyYjNCBhDAaGj7KSH7apGoSl5Ck9vAyiL2dv54c4tCX9dyswutZeK7+RgH91v+SsxKrp8= cloud@pritunl-demo",
		Principals:   []string{},
		Certificate:  "",
	},
	{
		Id:           utils.ObjectIdHex("688ab80d1793930f821f4f5f"),
		Name:         "demo",
		Comment:      "",
		Type:         "ssh_key",
		Organization: utils.ObjectIdHex("5a3245a50accad1a8a53bc82"),
		Roles:        []string{"instance"},
		Key:          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCtianWXiLSMZ5h53POvUd0v6oIEH7dw/KZryymxa1x+CEFSaeeACWNSalcWSx48D53v7S4OytXkqjjwdsxGjJ0jfYe9c06s3fCdp0to9Dz1Xk5jeW0ojZXit7Ta+4MQH7mbZkS5SOVWo09fvoUlRrYDD+dlpG1XmYqLVCY7Z2atTVArYSQ9xNQUbXU3TgljZ2yKPX25d20y50exJWxlJwEgXo8z2ZsUJO22KL23fLs3t5Dylj5uV7gD3lEKRe+v6VGESuY8QKFKLEuOy7F+xmZazkIxOixDT7bPOz1FHGzTFUUOanYD8F9zS8TZAHuW0B4yA3Uh90NQ+7mAqW1dcX1Qu78e8xuEtKzSVvLR02NMWqdD+/tebfb1QIB2ljz9PanHFpnT+Ht0RdONgNSqMIs4HORObXnNkvCYtMLtoy5acE9zhM4P+fyr6CMSGpiqhFxXFAT0x+Xws+KJjpO6kER/vmuOsTUCsxIPlfVntzeisVnMYomdeOipdQhMELt2yE= cloud@pritunl-demo",
		Principals:   []string{},
		Certificate:  "",
	},
}
