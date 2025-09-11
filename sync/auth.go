package sync

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/user"
	"github.com/sirupsen/logrus"
)

func authSync() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	coll := db.Users()
	opts := &options.CountOptions{}
	opts.SetLimit(1)

	count, err := coll.CountDocuments(
		db,
		&bson.M{
			"type": user.Local,
		},
		opts,
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	settings.Local.NoLocalAuth = count == 0

	return
}

func authRunner() {
	time.Sleep(1 * time.Second)

	for {
		time.Sleep(10 * time.Second)

		if constants.Shutdown {
			return
		}

		err := authSync()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("sync: Failed to sync authentication status")
		}
	}
}

func initAuth() {
	go authRunner()
}
