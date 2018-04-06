package data

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/minio/minio-go"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/utils"
	"time"
)

var (
	syncLock = utils.NewMultiTimeoutLock(1 * time.Minute)
)

func Sync(db *database.Database, store *storage.Storage) (err error) {
	if store.Endpoint == "" {
		return
	}

	lockId := syncLock.Lock(store.Id.Hex())
	defer syncLock.Unlock(store.Id.Hex(), lockId)

	client, err := minio.New(
		store.Endpoint, store.AccessKey, store.SecretKey, !store.Insecure)
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.New("storage: Failed to connect to storage"),
		}
		return
	}

	done := make(chan struct{})
	defer close(done)

	remoteKeys := set.NewSet()
	for object := range client.ListObjects(store.Bucket, "", true, done) {
		if object.Err != nil {
			err = &errortypes.RequestError{
				errors.New("storage: Failed to list objects"),
			}
			return
		}

		etag := image.GetEtag(object)
		remoteKeys.Add(object.Key)

		img := &image.Image{
			Storage:      store.Id,
			Key:          object.Key,
			Etag:         etag,
			Type:         store.Type,
			LastModified: object.LastModified,
		}
		err = img.Upsert(db)
		if err != nil {
			return
		}
	}

	localKeys, err := image.Distinct(db, store.Id)
	if err != nil {
		return
	}

	removeKeysSet := set.NewSet()
	for _, key := range localKeys {
		removeKeysSet.Add(key)
	}
	removeKeysSet.Subtract(remoteKeys)

	removeKeys := []string{}
	for key := range removeKeysSet.Iter() {
		removeKeys = append(removeKeys, key.(string))
	}

	err = image.RemoveKeys(db, store.Id, removeKeys)
	if err != nil {
		return
	}

	return
}
