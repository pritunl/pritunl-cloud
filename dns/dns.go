package dns

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/secret"
)

type Operation struct {
	Operation string
	Value     string
}

type Service interface {
	Connect(db *database.Database, secr *secret.Secret) (err error)
	DnsCommit(db *database.Database, domain, recordType string,
		ops []*Operation) (err error)
	DnsFind(db *database.Database, domain, recordType string) (
		vals []string, err error)
}
