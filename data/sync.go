package data

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

var (
	syncLock        = utils.NewMultiTimeoutLock(1 * time.Minute)
	clientTransport = &http.Transport{
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
		},
	}
	client = &http.Client{
		Transport: clientTransport,
		Timeout:   30 * time.Second,
	}
	clientLarge = &http.Client{
		Transport: clientTransport,
		Timeout:   30 * time.Minute,
	}
)

func getImagesS3(db *database.Database, store *storage.Storage) (
	images []*image.Image, err error) {

	client, err := minio.New(store.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(store.AccessKey, store.SecretKey, ""),
		Secure: !store.Insecure,
	})
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "storage: Failed to connect to storage"),
		}
		return
	}

	images = []*image.Image{}
	signedKeys := set.NewSet()
	for object := range client.ListObjects(
		context.Background(),
		store.Bucket, minio.ListObjectsOptions{
			Recursive: true,
		},
	) {
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

			img := &image.Image{
				Storage:      store.Id,
				Key:          object.Key,
				Firmware:     image.Uefi,
				Etag:         etag,
				Type:         store.Type,
				LastModified: object.LastModified,
			}

			if store.IsOracle() {
				obj, e := client.StatObject(context.Background(),
					store.Bucket, object.Key, minio.StatObjectOptions{})
				if e != nil {
					err = &errortypes.ReadError{
						errors.Wrap(e, "storage: Failed to stat object"),
					}
					return
				}

				img.StorageClass = storage.ParseStorageClass(obj)
			} else {
				img.StorageClass = storage.ParseStorageClass(object)
			}

			images = append(images, img)
		}
	}

	for _, img := range images {
		img.Signed = signedKeys.Contains(img.Key)
	}

	return
}

type Files struct {
	Version int `json:"version"`
	Files   []File
}

type File struct {
	Name         string    `json:"name"`
	Signed       bool      `json:"signed"`
	Hash         string    `json:"hash"`
	LastModified time.Time `json:"last_modified"`
}

func getImagesWeb(db *database.Database, store *storage.Storage) (
	images []*image.Image, err error) {

	u := store.GetWebUrl()
	u.Path += "/files.json"

	req, e := http.NewRequest("GET", u.String(), nil)
	if e != nil {
		err = &errortypes.RequestError{
			errors.Wrap(e, "data: Failed to file listing request"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl-cloud")

	resp, e := client.Do(req)
	if e != nil {
		err = &errortypes.RequestError{
			errors.Wrap(e, "data: File listing request error"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = &errortypes.RequestError{
			errors.Newf(
				"data: Bad status %d from file listing request",
				resp.StatusCode,
			),
		}
		return
	}

	filesData := &Files{}
	err = json.NewDecoder(resp.Body).Decode(filesData)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(
				err, "data: Failed to unmarshal file listing",
			),
		}
		return
	}

	images = []*image.Image{}
	for _, object := range filesData.Files {
		if strings.HasSuffix(object.Name, ".qcow2") {
			img := &image.Image{
				Storage:      store.Id,
				Key:          object.Name,
				Firmware:     image.Uefi,
				Etag:         object.Hash,
				Type:         storage.Web,
				Signed:       object.Signed,
				LastModified: object.LastModified,
			}

			images = append(images, img)
		}
	}

	return
}

func Sync(db *database.Database, store *storage.Storage) (err error) {
	if store.Endpoint == "" {
		return
	}

	lockId := syncLock.Lock(store.Id.Hex())
	defer syncLock.Unlock(store.Id.Hex(), lockId)

	var images []*image.Image

	if store.Type == storage.Web || store.Endpoint == "images.pritunl.com" {
		images, err = getImagesWeb(db, store)
		if err != nil {
			return
		}
	} else {
		images, err = getImagesS3(db, store)
		if err != nil {
			return
		}
	}

	remoteKeys := set.NewSet()
	for _, img := range images {
		remoteKeys.Add(img.Key)

		if img.Signed {
			if strings.Contains(img.Key, "_efi") ||
				strings.Contains(img.Key, "_uefi") {

				img.Firmware = image.Uefi
			} else {
				img.Firmware = image.Bios
			}
		}

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

	for keyInf := range removeKeysSet.Iter() {
		key := keyInf.(string)

		img, e := image.GetKey(db, store.Id, key)
		if e != nil {
			err = e
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}

		logrus.WithFields(logrus.Fields{
			"bucket": store.Bucket,
			"key":    img.Key,
		}).Info("data: Remote image deleted, removing local")

		err = img.Remove(db)
		if err != nil {
			return
		}
	}

	return
}
