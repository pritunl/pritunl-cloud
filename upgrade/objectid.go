package upgrade

import (
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
)

func objectIdUpgrade(db *database.Database) (err error) {
	nilObjectID := bson.NilObjectID

	coll := db.Alerts()
	_, err = coll.UpdateMany(db, bson.M{
		"organization": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"organization": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Authorities()
	_, err = coll.UpdateMany(db, bson.M{
		"organization": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"organization": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Balancers()
	_, err = coll.UpdateMany(db, bson.M{
		"organization": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"organization": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"datacenter": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"datacenter": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Certificates()
	_, err = coll.UpdateMany(db, bson.M{
		"organization": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"organization": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"acme_secret": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"acme_secret": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Deployments()
	_, err = coll.UpdateMany(db, bson.M{
		"datacenter": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"datacenter": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"zone": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"zone": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"node": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"node": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"instance": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"instance": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"image": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"image": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Disks()
	_, err = coll.UpdateMany(db, bson.M{
		"node": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"node": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"pool": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"pool": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"organization": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"organization": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"instance": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"instance": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"source_instance": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"source_instance": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"deployment": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"deployment": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"image": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"image": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"restore_image": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"restore_image": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Domains()
	_, err = coll.UpdateMany(db, bson.M{
		"organization": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"organization": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"lock_id": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"lock_id": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Firewalls()
	_, err = coll.UpdateMany(db, bson.M{
		"$or": []*bson.M{
			&bson.M{
				"organization": nil,
			},
			&bson.M{
				"organization": &bson.M{
					"$exists": false,
				},
			},
		},
	}, bson.M{
		"$set": bson.M{
			"organization": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Images()
	_, err = coll.UpdateMany(db, bson.M{
		"deployment": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"deployment": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Instances()
	_, err = coll.UpdateMany(db, bson.M{
		"disk_pool": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"disk_pool": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"node": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"node": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"shape": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"shape": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	_, err = coll.UpdateMany(db, bson.M{
		"deployment": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"deployment": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Plans()
	_, err = coll.UpdateMany(db, bson.M{
		"organization": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"organization": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Secrets()
	_, err = coll.UpdateMany(db, bson.M{
		"organization": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"organization": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	coll = db.Zones()
	_, err = coll.UpdateMany(db, bson.M{
		"datacenter": bson.M{
			"$exists": false,
		},
	}, bson.M{
		"$set": bson.M{
			"datacenter": nilObjectID,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
