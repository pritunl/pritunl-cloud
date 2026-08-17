package upgrade

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
)

type domainRecordUpgradeDoc struct {
	Id           bson.ObjectID `bson:"_id"`
	Domain       bson.ObjectID `bson:"domain"`
	Organization bson.ObjectID `bson:"organization"`
}

func domainRecordUpgrade(db *database.Database) (err error) {
	coll := db.Domains()

	domainOrgMap := make(map[bson.ObjectID]bson.ObjectID)

	cursor, err := coll.Find(db, bson.M{})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &domainRecordUpgradeDoc{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		domainOrgMap[doc.Id] = doc.Organization
	}
	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.DomainsRecords()
	cursor, err = coll.Find(db, bson.M{})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		doc := &domainRecordUpgradeDoc{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		organizationId, ok := domainOrgMap[doc.Domain]
		if !ok || doc.Organization == organizationId {
			continue
		}

		_, err = coll.UpdateOne(
			db,
			bson.M{
				"_id": doc.Id,
			},
			bson.M{
				"$set": bson.M{
					"organization": organizationId,
				},
			},
		)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
