package demo

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/utils"
)

var Secrets = []*secret.Secret{
	{
		Id:           utils.ObjectIdHex("67b89e8d4866ba90e6c459ba"),
		Name:         "cloudflare-pritunl-com",
		Comment:      "",
		Organization: bson.ObjectID{},
		Type:         "cloudflare",
		Key:          "a7kX9mN2vP8Q-4jL6wS3tR5Y-uH1gF7dZ0xC-vB8nM",
		Value:        "",
		Region:       "",
		PublicKey: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAz4K8Lm3QvR7WxN5YdE2P
jX9TpQ6HgM1wV0nS4KaF3ZcB8LrY5UvO2JmN7XsPqI1AgK8EoH3RdWzM9LfY2VtN
kP4QxGsJ7YnR8LwVmT3AqZ5HvK2NdP1XoS8JgR4LmW7YxQ3VnH5TsK9PpL2MdX8Rg
vJ3KqN5WxT1LsM4HgY7RdP8NqV2JmK5XwL3TsR8YgN4HxP1LdK9VwQ2MsT3XpR7Y
nL8KgJ5WdH3TmR9XsL2PqN7VxK4MgT3HdJ8YwP2LsK5RxT1NqM4JgY7PxR8WsL3T
mK9XwN2HgJ5YdL3RsP8VqT2MxK4NhR3JdY8WwL2TsM5QxN1PqK4YgJ7RxP8VsT3M
PwIDAQAB
-----END PUBLIC KEY-----`,
		Data: "",
	},
}
