package data

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/errors"
	"github.com/minio/minio-go"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
	"io"
	"os"
	"path"
)

func getImage(db *database.Database, img *image.Image,
	pth string) (err error) {

	store, err := storage.Get(db, img.Storage)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"id":         img.Id.Hex(),
		"storage_id": store.Id.Hex(),
		"key":        img.Key,
	}).Info("data: Downloading image")

	client, err := minio.New(
		store.Endpoint, store.AccessKey, store.SecretKey, !store.Insecure)
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "data: Failed to connect to storage"),
		}
		return
	}

	object, err := client.GetObject(store.Bucket,
		img.Key, minio.GetObjectOptions{})
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "data: Failed to read image"),
		}
		return
	}
	defer object.Close()

	cacheFile, err := os.OpenFile(pth,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "data: Failed to create image cache file"),
		}
		return
	}
	defer cacheFile.Close()

	stat, err := object.Stat()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "data: Failed to stat image"),
		}
		return
	}

	_, err = io.CopyN(cacheFile, object, stat.Size)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "data: Failed to copy image cache file"),
		}
		return
	}

	return
}

func WriteImage(db *database.Database, imgId bson.ObjectId, pth string) (
	err error) {

	img, err := image.Get(db, imgId)
	if err != nil {
		return
	}

	cacheDir := node.Self.GetCachePath()

	cachePth := path.Join(
		cacheDir,
		fmt.Sprintf("%s_%s", img.Id.Hex(), img.Etag),
	)

	err = utils.ExistsMkdir(cacheDir, 0755)
	if err != nil {
		return
	}

	exists, err := utils.Exists(cachePth)
	if err != nil {
		return
	}

	if !exists {
		err = getImage(db, img, cachePth)
		if err != nil {
			return
		}
	}

	err = utils.Exec("", "cp", cachePth, pth)
	if err != nil {
		return
	}

	return
}
