package demo

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Certificates = []*certificate.Certificate{
	{
		Id:           utils.ObjectIdHex("67b89ef24866ba90e6c459e8"),
		Name:         "cloud-pritunl-com",
		Comment:      "",
		Organization: bson.ObjectID{},
		Type:         "lets_encrypt",
		Key: `-----BEGIN RSA PRIVATE KEY-----
MIIJKQIBAAKCAgEAx9Y3Lk2AwV6ap7L/Sx9XC5mXaUf8hvMmDbLBqDZ1Y7xKJM2h
zQ8Xm1rK9q0wzQC6qiL6xHmTpKWTzNVzGsQdM3/qNPLNA7W8PIYCzjkSe5X1YktY
vxldBxYxPRJxXk5S9P8dFYVmFFKF2bvJ5pSMLq9w3z3nTm3TQtRPqWx2Vk3DqV2D
QKmNtqJnhVqYvVKa3QpLLwz8xKqB1sPXLr4XqQ3bz3fLjLxPmYV5WxLhgdKLYZTv
YxQPLPTJkX3Pw4XD4Qs4CrKLW5bYsqYKQ7kKDXgJmTxYzZLjZKf4vSqLxqV5bDPY
rR2YxQ9TKLkYKVMpNtY5J9X2fWzyPSvXqXZfVx7D8xJzDY8YKPLXmvxKQZxLJxSx
zxHQzYKJpX3YmVfqYYmfYxXYzLmYxDzSxXqLvKxVqXxQDsPxQVKfKqQx5KvxsVqD
-----END RSA PRIVATE KEY-----`,
		Certificate: `-----BEGIN CERTIFICATE-----
MIIGGTCCBQGgAwIBAgISBXx9YmN2KQm9g3Y5XmKbvx9YMA0GCSqGSIb3DQEBCwUA
MDMxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MQwwCgYDVQQD
EwNSMTEwHhcNMjUwODA4MDY0NzI3WhcNMjUxMTA2MDY0NzI2WjAcMRowGAYDVQQD
ExFjbG91ZC5wcml0dW5sLnJlZDCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoC
ggIBAMfWNy5NgMFemqey/0sfVwuZl2lH/IbzJg2ywag2dWO8SiTNoc0PF5tayvat
MM0AuKoi+sR5k6Slk8zVcxrEHTN/6jTyzQO1vDyGAs45EnuV9WJLWL8ZXQcWMT0S
cV5OUvT/HRWFZhRShdn5iQ2Sry6vcN8950Dt00LUT6lsdlZNw6ldg0CpjbaiZ4Va
mL1Smt0KSy8M/MSqgdbD1y6+F6kN2893y4y8T5mFeVsS4YHSi2GU72MUDyz0yZF9
z8OFw+ELOAqyi1uW2LKmCkO5Cg14CZk8WM2S42Sn+L0qi8aleWwz2K0dmMUPUyi5
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIFBjCCAu6gAwIBAgIRAIp9PhPWLzDvI4a9KQdrNPgwDQYJKoZIhvcNAQELBQAw
TzELMAkGA1UEBhMCVVMxKTAnBgNVBAoTIEludGVybmV0IFNlY3VyaXR5IFJlc2Vh
cmNoIEdyb3VwMRUwEwYDVQQDEwxJU1JHIFJvb3QgWDEwHhcNMjQwMzEzMDAwMDAw
WhcNMjcwMzEyMjM1OTU5WjAzMQswCQYDVQQGEwJVUzEWMBQGA1UEChMNTGV0J3Mg
RW5jcnlwdDEMMAoGA1UEAxMDUjExMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAuoe8XBsAOcvKCs3UZxD5ATylTqVhyybKUvsVAbe5KPUoHu0nsyQYOWcJ
DAjs4DqwO3cOvfPlOVRBDE6uQdaZdN5R2+97/1i9qLcT9t4x1fJyyXJqC4N0lZxG
AGQUmfOx2SLZzaiSqhwmej/+71gFewiVgdtxD4774zEJuwm+UE1fj5F2PVqdnoPy
-----END CERTIFICATE-----`,
		Info: &certificate.Info{
			Hash:         "bba8a3941280c8466a6a2a723cc06f26",
			SignatureAlg: "SHA256-RSA",
			PublicKeyAlg: "RSA",
			Issuer:       "R11",
			IssuedOn:     time.Now(),
			ExpiresOn:    time.Now().Add(2160 * time.Hour),
			DnsNames: []string{
				"cloud.pritunl.com",
				"user.cloud.pritunl.com",
			},
		},
		AcmeDomains: []string{
			"cloud.pritunl.com",
			"user.cloud.pritunl.com",
		},
		AcmeType:   "acme_dns",
		AcmeAuth:   "acme_cloudflare",
		AcmeSecret: utils.ObjectIdHex("67b89e8d4866ba90e6c459ba"),
	},
}
