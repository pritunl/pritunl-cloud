package journal

import (
	"github.com/pritunl/pritunl-cloud/database"
)

type KindGenerator interface {
	GetKind(db *database.Database, key string) (kind int32, err error)
}
