package secret

import (
	"bytes"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

func Get(db *database.Database, secrId primitive.ObjectID) (
	secr *Secret, err error) {

	coll := db.Secrets()
	secr = &Secret{}

	err = coll.FindOneId(secrId, secr)
	if err != nil {
		return
	}

	return
}

func GetOne(db *database.Database, query *bson.M) (secr *Secret, err error) {
	coll := db.Secrets()
	secr = &Secret{}

	err = coll.FindOne(db, query).Decode(secr)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetOrg(db *database.Database, orgId, secrId primitive.ObjectID) (
	secr *Secret, err error) {

	coll := db.Secrets()
	secr = &Secret{}

	err = coll.FindOne(db, &bson.M{
		"_id":          secrId,
		"organization": orgId,
	}).Decode(secr)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAll(db *database.Database, query *bson.M) (
	secrs []*Secret, err error) {

	coll := db.Secrets()
	secrs = []*Secret{}

	cursor, err := coll.Find(db, query)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		secr := &Secret{}
		err = cursor.Decode(secr)
		if err != nil {
			return
		}

		secrs = append(secrs, secr)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllOrg(db *database.Database, orgId primitive.ObjectID) (
	secrs []*Secret, err error) {

	coll := db.Secrets()
	secrs = []*Secret{}

	cursor, err := coll.Find(db, &bson.M{
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		secr := &Secret{}
		err = cursor.Decode(secr)
		if err != nil {
			return
		}

		secrs = append(secrs, secr)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetAllPaged(db *database.Database, query *bson.M,
	page, pageCount int64) (secrs []*Secret, count int64, err error) {

	coll := db.Secrets()
	secrs = []*Secret{}

	if len(*query) == 0 {
		count, err = coll.EstimatedDocumentCount(db)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	} else {
		count, err = coll.CountDocuments(db, query)
		if err != nil {
			err = database.ParseError(err)
			return
		}
	}

	maxPage := count / pageCount
	if count == pageCount {
		maxPage = 0
	}
	page = utils.Min64(page, maxPage)
	skip := utils.Min64(page*pageCount, count)

	cursor, err := coll.Find(
		db,
		query,
		&options.FindOptions{
			Sort: &bson.D{
				{"name", 1},
			},
			Skip:  &skip,
			Limit: &pageCount,
		},
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		secr := &Secret{}
		err = cursor.Decode(secr)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		secrs = append(secrs, secr)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func ExistsOrg(db *database.Database, orgId, secrId primitive.ObjectID) (
	exists bool, err error) {

	coll := db.Secrets()
	n, err := coll.CountDocuments(
		db,
		&bson.M{
			"_id":          secrId,
			"organization": orgId,
		},
	)
	if err != nil {
		return
	}

	if n > 0 {
		exists = true
	}

	return
}

func Remove(db *database.Database, secrId primitive.ObjectID) (err error) {
	coll := db.Secrets()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": secrId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveOrg(db *database.Database, orgId, secrId primitive.ObjectID) (
	err error) {

	coll := db.Secrets()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id":          secrId,
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		switch err.(type) {
		case *database.NotFoundError:
			err = nil
		default:
			return
		}
	}

	return
}

func RemoveMulti(db *database.Database, secrIds []primitive.ObjectID) (
	err error) {
	coll := db.Secrets()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": secrIds,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func RemoveMultiOrg(db *database.Database, orgId primitive.ObjectID,
	secrIds []primitive.ObjectID) (err error) {

	coll := db.Secrets()

	_, err = coll.DeleteMany(db, &bson.M{
		"_id": &bson.M{
			"$in": secrIds,
		},
		"organization": orgId,
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func loadPrivateKey(secr *Secret) (
	key *rsa.PrivateKey, fingerprint string, err error) {

	block, _ := pem.Decode([]byte(secr.PrivateKey))
	if block == nil {
		err = &errortypes.ParseError{
			errors.New("secret: Failed to decode private key"),
		}
		return
	}

	if block.Type != "RSA PRIVATE KEY" {
		err = &errortypes.ParseError{
			errors.New("secret: Invalid private key type"),
		}
		return
	}

	key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "secret: Failed to parse rsa key"),
		}
		return
	}

	pubKey, err := x509.MarshalPKIXPublicKey(key.Public())
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "secret: Failed to marshal public key"),
		}
		return
	}

	keyHash := md5.New()
	keyHash.Write(pubKey)
	fingerprint = fmt.Sprintf("%x", keyHash.Sum(nil))
	fingerprintBuf := bytes.Buffer{}

	for i, run := range fingerprint {
		fingerprintBuf.WriteRune(run)
		if i%2 == 1 && i != len(fingerprint)-1 {
			fingerprintBuf.WriteRune(':')
		}
	}
	fingerprint = fingerprintBuf.String()

	return
}
