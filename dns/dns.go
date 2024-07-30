package dns

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/secret"
)

type Service interface {
	Connect(db *database.Database, secr *secret.Secret) (err error)
	DnsTxtGet(db *database.Database, domain string) (vals []string, err error)
	DnsTxtUpsert(db *database.Database, domain, val string) (err error)
	DnsTxtDelete(db *database.Database, domain, val string) (err error)
	DnsAUpsert(db *database.Database, domain, val string) (err error)
	DnsADelete(db *database.Database, domain, val string) (err error)
	DnsAAAAUpsert(db *database.Database, domain, val string) (err error)
	DnsAAAADelete(db *database.Database, domain, val string) (err error)
}
