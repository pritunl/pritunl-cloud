package data

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/errors"
	"github.com/minio/minio-go"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/zone"
	"golang.org/x/crypto/openpgp"
	"gopkg.in/mgo.v2/bson"
	"os"
	"path"
	"strings"
	"time"
)

var (
	imageLock        = utils.NewMultiTimeoutLock(10 * time.Minute)
	backingImageLock = utils.NewMultiTimeoutLock(5 * time.Minute)
)

func getImage(db *database.Database, img *image.Image,
	pth string) (err error) {

	if imageLock.Locked(pth) {
		logrus.WithFields(logrus.Fields{
			"id":   img.Id.Hex(),
			"key":  img.Key,
			"path": pth,
		}).Info("data: Waiting for image")
	}

	lockId := imageLock.Lock(pth)
	defer imageLock.Unlock(pth, lockId)

	exists, err := utils.Exists(pth)
	if err != nil {
		return
	}

	if exists {
		return
	}

	tmpPth := paths.GetImageTempPath()

	store, err := storage.Get(db, img.Storage)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"id":         img.Id.Hex(),
		"storage_id": store.Id.Hex(),
		"key":        img.Key,
		"path":       pth,
	}).Info("data: Downloading image")

	client, err := minio.New(
		store.Endpoint, store.AccessKey, store.SecretKey, !store.Insecure)
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "data: Failed to connect to storage"),
		}
		return
	}

	err = client.FGetObject(store.Bucket,
		img.Key, tmpPth, minio.GetObjectOptions{})
	if err != nil {
		os.Remove(tmpPth)

		err = &errortypes.ReadError{
			errors.Wrap(err, "data: Failed to download image"),
		}
		return
	}

	if strings.Contains(store.Endpoint, "images.pritunl.com") {
		sigPth := tmpPth + ".sig"
		defer os.Remove(sigPth)

		err = client.FGetObject(store.Bucket,
			img.Key+".sig", sigPth, minio.GetObjectOptions{})
		if err != nil {
			os.Remove(tmpPth)

			err = &errortypes.ReadError{
				errors.Wrap(err, "data: Failed to download image signature"),
			}
			return
		}

		signature, e := os.Open(sigPth)
		if e != nil {
			os.Remove(tmpPth)

			err = &errortypes.ReadError{
				errors.Wrap(e, "data: Failed to open image signature"),
			}
			return
		}
		defer signature.Close()

		tmpImg, e := os.Open(tmpPth)
		if e != nil {
			os.Remove(tmpPth)

			err = &errortypes.ReadError{
				errors.Wrap(e, "data: Failed to open image"),
			}
			return
		}
		defer tmpImg.Close()

		keyring, e := openpgp.ReadArmoredKeyRing(
			strings.NewReader(constants.PritunlKeyring))
		if e != nil {
			os.Remove(tmpPth)

			err = &errortypes.ParseError{
				errors.Wrap(e, "data: Failed to parse Pritunl keyring"),
			}
			return
		}

		entity, e := openpgp.CheckArmoredDetachedSignature(
			keyring, tmpImg, signature)
		if e != nil || entity == nil {
			os.Remove(tmpPth)

			err = &errortypes.VerificationError{
				errors.Wrap(e, "data: Image signature verification failed"),
			}
			return
		}

		logrus.WithFields(logrus.Fields{
			"id":         img.Id.Hex(),
			"storage_id": store.Id.Hex(),
			"key":        img.Key,
		}).Info("data: Image signature successfully validated")
	}

	err = utils.Exec("", "mv", tmpPth, pth)
	if err != nil {
		return
	}

	return
}

func copyBackingImage(imagePth, backingImagePth string) (err error) {
	lockId := backingImageLock.Lock(backingImagePth)
	defer backingImageLock.Unlock(backingImagePth, lockId)

	exists, err := utils.Exists(backingImagePth)
	if err != nil {
		return
	}

	if exists {
		return
	}

	err = utils.Exec("", "cp", imagePth, backingImagePth)
	if err != nil {
		return
	}

	return
}

func WriteImage(db *database.Database, imgId, dskId bson.ObjectId,
	size int) (err error) {

	diskPath := paths.GetDiskPath(dskId)
	diskTempPath := paths.GetDiskTempPath()
	disksPath := paths.GetDisksPath()

	err = utils.ExistsMkdir(disksPath, 0755)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(paths.GetTempPath(), 0755)
	if err != nil {
		return
	}

	img, err := image.Get(db, imgId)
	if err != nil {
		return
	}

	if img.Type == storage.Public {
		cacheDir := node.Self.GetCachePath()

		imagePth := path.Join(
			cacheDir,
			fmt.Sprintf("image-%s-%s", img.Id.Hex(), img.Etag),
		)

		err = utils.ExistsMkdir(cacheDir, 0755)
		if err != nil {
			return
		}

		err = getImage(db, img, imagePth)
		if err != nil {
			return
		}

		exists, e := utils.Exists(diskPath)
		if e != nil {
			err = e
			return
		}

		if exists {
			logrus.WithFields(logrus.Fields{
				"image_id":   img.Id.Hex(),
				"image_type": img.Type,
				"disk_id":    dskId.Hex(),
				"key":        img.Key,
				"path":       diskPath,
			}).Error("data: Blocking disk image overwrite")

			err = &errortypes.WriteError{
				errors.Wrap(err, "data: Image already exists"),
			}
			return
		}

		utils.Exec("", "touch", imagePth)

		err = utils.Exec("", "cp", imagePth, diskTempPath)
		if err != nil {
			return
		}

		if size > 10 {
			_, err = utils.ExecCombinedOutputLogged(nil, "qemu-img",
				"resize", diskTempPath, fmt.Sprintf("%dG", size))
			if err != nil {
				return
			}
		}

		err = utils.Exec("", "mv", diskTempPath, diskPath)
		if err != nil {
			return
		}
	} else {
		exists, e := utils.Exists(diskPath)
		if e != nil {
			err = e
			return
		}

		if exists {
			logrus.WithFields(logrus.Fields{
				"image_id":   img.Id.Hex(),
				"image_type": img.Type,
				"disk_id":    dskId.Hex(),
				"key":        img.Key,
				"path":       diskPath,
			}).Error("data: Blocking disk image overwrite")

			err = &errortypes.WriteError{
				errors.Wrap(err, "data: Image already exists"),
			}
			return
		}

		err = getImage(db, img, diskTempPath)
		if err != nil {
			return
		}

		if size > 10 {
			_, err = utils.ExecCombinedOutputLogged(nil, "qemu-img",
				"resize", diskTempPath, fmt.Sprintf("%dG", size))
			if err != nil {
				return
			}
		}

		err = utils.Exec("", "mv", diskTempPath, diskPath)
		if err != nil {
			return
		}
	}

	return
}

func DeleteImage(db *database.Database, imgId bson.ObjectId) (err error) {
	img, err := image.Get(db, imgId)
	if err != nil {
		return
	}

	if img.Type == storage.Public {
		return
	}

	store, err := storage.Get(db, img.Storage)
	if err != nil {
		return
	}

	client, err := minio.New(
		store.Endpoint, store.AccessKey, store.SecretKey, !store.Insecure)
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "data: Failed to connect to storage"),
		}
		return
	}

	err = client.RemoveObject(store.Bucket, img.Key)
	if err != nil {
		return
	}

	err = image.Remove(db, img.Id)
	if err != nil {
		return
	}

	return
}

func DeleteImages(db *database.Database, imgIds []bson.ObjectId) (err error) {
	for _, imgId := range imgIds {
		err = DeleteImage(db, imgId)
		if err != nil {
			return
		}
	}

	return
}

func DeleteImageOrg(db *database.Database, orgId, imgId bson.ObjectId) (
	err error) {

	img, err := image.GetOrg(db, orgId, imgId)
	if err != nil {
		return
	}

	if img.Type == storage.Public {
		return
	}

	store, err := storage.Get(db, img.Storage)
	if err != nil {
		return
	}

	client, err := minio.New(
		store.Endpoint, store.AccessKey, store.SecretKey, !store.Insecure)
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "data: Failed to connect to storage"),
		}
		return
	}

	err = client.RemoveObject(store.Bucket, img.Key)
	if err != nil {
		return
	}

	err = image.Remove(db, img.Id)
	if err != nil {
		return
	}

	return
}

func DeleteImagesOrg(db *database.Database, orgId bson.ObjectId,
	imgIds []bson.ObjectId) (err error) {

	for _, imgId := range imgIds {
		err = DeleteImageOrg(db, orgId, imgId)
		if err != nil {
			return
		}
	}

	return
}

func CreateSnapshot(db *database.Database, dsk *disk.Disk) (err error) {
	dskPth := paths.GetDiskPath(dsk.Id)
	cacheDir := node.Self.GetCachePath()

	logrus.WithFields(logrus.Fields{
		"disk_id":     dsk.Id.Hex(),
		"source_path": dskPth,
	}).Info("data: Creating disk snapshot")

	nde, err := node.Get(db, dsk.Node)
	if err != nil {
		return
	}

	zne, err := zone.Get(db, nde.Zone)
	if err != nil {
		return
	}

	dc, err := datacenter.Get(db, zne.Datacenter)
	if err != nil {
		return
	}

	if dc.PrivateStorage == "" {
		logrus.WithFields(logrus.Fields{
			"disk_id": dsk.Id.Hex(),
		}).Error("data: Cannot snapshot disk without private storage")
		return
	}

	store, err := storage.Get(db, dc.PrivateStorage)
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
			logrus.WithFields(logrus.Fields{
				"disk_id": dsk.Id.Hex(),
			}).Error("data: Cannot snapshot disk without private storage")
		}
		return
	}

	imgId := bson.NewObjectId()
	tmpPath := path.Join(cacheDir,
		fmt.Sprintf("snapshot-%s", imgId.Hex()))
	img := &image.Image{
		Id: imgId,
		Name: fmt.Sprintf("%s-%s", dsk.Name,
			time.Now().Format("2006-01-02T15:04:05")),
		Organization: dsk.Organization,
		Type:         storage.Private,
		Storage:      store.Id,
		Key:          fmt.Sprintf("snapshot/%s.qcow2", imgId.Hex()),
	}

	defer utils.Remove(tmpPath)
	err = utils.Exec("", "qemu-img", "convert", "-f", "qcow2",
		"-O", "qcow2", "-c", dskPth, tmpPath)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"disk_id":     dsk.Id.Hex(),
		"source_path": dskPth,
		"storage_id":  store.Id.Hex(),
		"object_key":  img.Key,
	}).Info("data: Uploading disk snapshot")

	client, err := minio.New(
		store.Endpoint, store.AccessKey, store.SecretKey, !store.Insecure)
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "data: Failed to connect to storage"),
		}
		return
	}

	_, err = client.FPutObject(store.Bucket, img.Key, tmpPath,
		minio.PutObjectOptions{})
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "data: Failed to write object"),
		}
		return
	}

	obj, err := client.StatObject(store.Bucket, img.Key,
		minio.StatObjectOptions{})
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "data: Failed to stat object"),
		}
		return
	}

	img.Etag = image.GetEtag(obj)
	img.LastModified = obj.LastModified

	err = img.Insert(db)
	if err != nil {
		client.RemoveObject(store.Bucket, img.Key)
		return
	}

	event.PublishDispatch(db, "image.change")

	return
}
