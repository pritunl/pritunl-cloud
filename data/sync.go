package data

import (
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/minio/minio-go"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/utils"
	"strings"
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
			errors.Wrap(err, "storage: Failed to connect to storage"),
		}
		return
	}

	done := make(chan struct{})
	defer close(done)

	images := []*image.Image{}
	signedKeys := set.NewSet()
	remoteKeys := set.NewSet()
	for object := range client.ListObjects(store.Bucket, "", true, done) {
		if object.Err != nil {
			err = &errortypes.RequestError{
				errors.Wrap(object.Err, "storage: Failed to list objects"),
			}
			return
		}

		if strings.HasSuffix(object.Key, ".qcow2.sig") {
			signedKeys.Add(strings.TrimRight(object.Key, ".sig"))
		} else if strings.HasSuffix(object.Key, ".qcow2") {
			etag := image.GetEtag(object)
			remoteKeys.Add(object.Key)

			img := &image.Image{
				Storage:      store.Id,
				Key:          object.Key,
				Etag:         etag,
				Type:         store.Type,
				LastModified: object.LastModified,
				StorageClass: storage.ParseStorageClass(object.StorageClass),
			}

			images = append(images, img)
		}
	}

	for _, img := range images {
		img.Signed = signedKeys.Contains(img.Key)

		err = img.Sync(db)
		if err != nil {
			if _, ok := err.(*image.LostImageError); ok {
				logrus.WithFields(logrus.Fields{
					"bucket": store.Bucket,
					"key":    img.Key,
				}).Error("data: Ignoring lost image")
			} else {
				return
			}
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
