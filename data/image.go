package data

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/qmp"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/vm"
	"github.com/pritunl/pritunl-cloud/zone"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/openpgp"
)

var (
	imageLock        = utils.NewMultiTimeoutLock(10 * time.Minute)
	backingImageLock = utils.NewMultiTimeoutLock(5 * time.Minute)
)

func getImage(db *database.Database, img *image.Image,
	pth string) (err error) {

	if imageLock.Locked(pth) {
		logrus.WithFields(logrus.Fields{
			"image_id": img.Id.Hex(),
			"key":      img.Key,
			"path":     pth,
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
		"image_id":   img.Id.Hex(),
		"storage_id": store.Id.Hex(),
		"key":        img.Key,
		"path":       pth,
	}).Info("data: Downloading image")

	client, err := minio.New(store.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			store.AccessKey,
			store.SecretKey,
			"",
		),
		Secure: !store.Insecure,
	})
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "data: Failed to connect to storage"),
		}
		return
	}

	err = client.FGetObject(context.Background(), store.Bucket,
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

		err = client.FGetObject(context.Background(), store.Bucket,
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

func WriteImage(db *database.Database, imgId, dskId primitive.ObjectID,
	size int, backingImage bool) (newSize int, backingImageName string,
	err error) {

	diskPath := paths.GetDiskPath(dskId)
	diskTempPath := paths.GetDiskTempPath()
	disksPath := paths.GetDisksPath()
	backingPath := paths.GetBackingPath()

	err = utils.ExistsMkdir(disksPath, 0755)
	if err != nil {
		return
	}

	err = utils.ExistsMkdir(backingPath, 0755)
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

	largeBase := strings.Contains(img.Key, "fedora")

	backingImagePth := path.Join(
		backingPath,
		fmt.Sprintf("image-%s-%s", img.Id.Hex(), img.Etag),
	)

	backingImageExists := false
	if backingImage {
		backingImageName = fmt.Sprintf("%s-%s", img.Id.Hex(), img.Etag)

		backingImageExists, err = utils.Exists(backingImagePth)
		if err != nil {
			return
		}
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

		if !backingImageExists {
			err = getImage(db, img, imagePth)
			if err != nil {
				return
			}
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

		if backingImage {
			err = copyBackingImage(imagePth, backingImagePth)
			if err != nil {
				return
			}

			utils.Exec("", "touch", backingImagePth)

			err = utils.Chmod(backingImagePth, 0644)
			if err != nil {
				return
			}

			if largeBase && size < 16 {
				size = 16
				newSize = 16
			} else if !largeBase && size < 10 {
				size = 10
				newSize = 10
			}

			_, err = utils.ExecCombinedOutputLogged(nil, "qemu-img",
				"create", "-f", "qcow2", "-F", "qcow2",
				"-o", fmt.Sprintf("backing_file=%s", backingImagePth),
				diskTempPath,
				fmt.Sprintf("%dG", size))
			if err != nil {
				return
			}
		} else {
			err = utils.Exec("", "cp", imagePth, diskTempPath)
			if err != nil {
				return
			}

			if largeBase && size < 16 {
				size = 16
				newSize = 16
			}

			if (largeBase && size > 16) || (!largeBase && size > 10) {
				_, err = utils.ExecCombinedOutputLogged(nil, "qemu-img",
					"resize", diskTempPath, fmt.Sprintf("%dG", size))
				if err != nil {
					return
				}
			}
		}

		err = utils.Chmod(diskTempPath, 0600)
		if err != nil {
			return
		}

		err = utils.Exec("", "mv", diskTempPath, diskPath)
		if err != nil {
			return
		}
	} else {
		if backingImage {
			err = getImage(db, img, backingImagePth)
			if err != nil {
				return
			}
		} else {
			err = getImage(db, img, diskTempPath)
			if err != nil {
				return
			}
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

		if backingImage {
			utils.Exec("", "touch", backingImagePth)

			err = utils.Chmod(backingImagePth, 0644)
			if err != nil {
				return
			}

			if largeBase && size < 16 {
				size = 16
				newSize = 16
			} else if !largeBase && size < 10 {
				size = 10
				newSize = 10
			}

			_, err = utils.ExecCombinedOutputLogged(nil, "qemu-img",
				"create", "-f", "qcow2", "-F", "qcow2",
				"-o", fmt.Sprintf("backing_file=%s", backingImagePth),
				diskTempPath,
				fmt.Sprintf("%dG", size))
			if err != nil {
				return
			}
		} else {
			if largeBase && size < 16 {
				size = 16
				newSize = 16
			}

			if (largeBase && size > 16) || (!largeBase && size > 10) {
				_, err = utils.ExecCombinedOutputLogged(nil, "qemu-img",
					"resize", diskTempPath, fmt.Sprintf("%dG", size))
				if err != nil {
					return
				}
			}
		}

		err = utils.Exec("", "mv", diskTempPath, diskPath)
		if err != nil {
			return
		}
	}

	return
}

func DeleteImage(db *database.Database, imgId primitive.ObjectID) (err error) {
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

	client, err := minio.New(store.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(store.AccessKey, store.SecretKey, ""),
		Secure: !store.Insecure,
	})
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "data: Failed to connect to storage"),
		}
		return
	}

	err = client.RemoveObject(context.Background(),
		store.Bucket, img.Key, minio.RemoveObjectOptions{})
	if err != nil {
		return
	}

	err = image.Remove(db, img.Id)
	if err != nil {
		return
	}

	return
}

func DeleteImages(db *database.Database, imgIds []primitive.ObjectID) (
	err error) {

	for _, imgId := range imgIds {
		err = DeleteImage(db, imgId)
		if err != nil {
			return
		}
	}

	return
}

func DeleteImageOrg(db *database.Database, orgId, imgId primitive.ObjectID) (
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

	client, err := minio.New(store.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(store.AccessKey, store.SecretKey, ""),
		Secure: !store.Insecure,
	})
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "data: Failed to connect to storage"),
		}
		return
	}

	err = client.RemoveObject(context.Background(),
		store.Bucket, img.Key, minio.RemoveObjectOptions{})
	if err != nil {
		return
	}

	err = image.Remove(db, img.Id)
	if err != nil {
		return
	}

	return
}

func DeleteImagesOrg(db *database.Database, orgId primitive.ObjectID,
	imgIds []primitive.ObjectID) (err error) {

	for _, imgId := range imgIds {
		err = DeleteImageOrg(db, orgId, imgId)
		if err != nil {
			return
		}
	}

	return
}

func CreateSnapshot(db *database.Database, dsk *disk.Disk,
	virt *vm.VirtualMachine) (err error) {

	dskPth := paths.GetDiskPath(dsk.Id)
	cacheDir := node.Self.GetCachePath()

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

	if dc.PrivateStorage.IsZero() {
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

	logrus.WithFields(logrus.Fields{
		"disk_id":    dsk.Id.Hex(),
		"storage_id": store.Id.Hex(),
		"disk_path":  dskPth,
	}).Info("data: Creating disk snapshot")

	err = utils.ExistsMkdir(cacheDir, 0755)
	if err != nil {
		return
	}

	imgId := primitive.NewObjectID()
	tmpPath := path.Join(cacheDir,
		fmt.Sprintf("snapshot-%s", imgId.Hex()))
	img := &image.Image{
		Id: imgId,
		Name: fmt.Sprintf("%s-%s", dsk.Name,
			time.Now().Format("2006-01-02T15:04:05")),
		Organization: dsk.Organization,
		Type:         storage.Private,
		Firmware:     image.Unknown,
		Storage:      store.Id,
		Key:          fmt.Sprintf("snapshot/%s.qcow2", imgId.Hex()),
	}

	defer utils.Remove(tmpPath)

	available := false
	if virt != nil && virt.Running() {
		err = qmp.BackupDisk(virt.Id, dsk, tmpPath)
		if err != nil {
			if _, ok := err.(*qmp.DiskNotFound); ok {
				err = nil
			} else {
				return
			}
		} else {
			available = true
		}
	}

	if !available {
		err = utils.Exec("", "cp", dskPth, tmpPath)
		if err != nil {
			return
		}
	}

	err = utils.Chmod(tmpPath, 0600)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"disk_id":    dsk.Id.Hex(),
		"disk_path":  dskPth,
		"storage_id": store.Id.Hex(),
		"object_key": img.Key,
	}).Info("data: Uploading disk snapshot")

	client, err := minio.New(store.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(store.AccessKey, store.SecretKey, ""),
		Secure: !store.Insecure,
	})
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "data: Failed to connect to storage"),
		}
		return
	}

	putOpts := minio.PutObjectOptions{}
	storageClass := storage.FormatStorageClass(dc.PrivateStorageClass)
	if storageClass != "" {
		putOpts.StorageClass = storageClass
	}

	_, err = client.FPutObject(context.Background(),
		store.Bucket, img.Key, tmpPath, putOpts)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "data: Failed to write object"),
		}

		return
	}

	time.Sleep(3 * time.Second)

	obj, err := client.StatObject(context.Background(),
		store.Bucket, img.Key, minio.StatObjectOptions{})
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "data: Failed to stat object"),
		}
		return
	}

	img.Etag = image.GetEtag(obj)
	img.LastModified = obj.LastModified

	if store.IsOracle() {
		img.StorageClass = storage.ParseStorageClass(obj)
	} else {
		img.StorageClass = dc.BackupStorageClass
	}

	err = img.Upsert(db)
	if err != nil {
		return
	}

	event.PublishDispatch(db, "image.change")

	return
}

func CreateBackup(db *database.Database, dsk *disk.Disk,
	virt *vm.VirtualMachine) (err error) {

	dskPth := paths.GetDiskPath(dsk.Id)
	cacheDir := node.Self.GetCachePath()

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

	if dc.BackupStorage.IsZero() {
		logrus.WithFields(logrus.Fields{
			"disk_id": dsk.Id.Hex(),
		}).Error("data: Cannot backup disk without backup storage")
		return
	}

	if dsk.BackingImage != "" {
		logrus.WithFields(logrus.Fields{
			"disk_id": dsk.Id.Hex(),
		}).Error("data: Cannot backup disk with backing image")
		return
	}

	store, err := storage.Get(db, dc.BackupStorage)
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
			logrus.WithFields(logrus.Fields{
				"disk_id": dsk.Id.Hex(),
			}).Error("data: Cannot backup disk without backup storage")
		}
		return
	}

	logrus.WithFields(logrus.Fields{
		"disk_id":    dsk.Id.Hex(),
		"storage_id": store.Id.Hex(),
		"disk_path":  dskPth,
	}).Info("data: Creating disk backup")

	err = utils.ExistsMkdir(cacheDir, 0755)
	if err != nil {
		return
	}

	imgId := primitive.NewObjectID()
	tmpPath := path.Join(cacheDir,
		fmt.Sprintf("backup-%s", imgId.Hex()))
	img := &image.Image{
		Id:   imgId,
		Disk: dsk.Id,
		Name: fmt.Sprintf("%s-%s", dsk.Name,
			time.Now().Format("2006-01-02T15:04:05")),
		Organization: dsk.Organization,
		Type:         storage.Private,
		Firmware:     image.Unknown,
		Storage:      store.Id,
		Key:          fmt.Sprintf("backup/%s.qcow2", imgId.Hex()),
	}

	defer utils.Remove(tmpPath)

	available := false
	if virt != nil && virt.Running() {
		err = qmp.BackupDisk(virt.Id, dsk, tmpPath)
		if err != nil {
			if _, ok := err.(*qmp.DiskNotFound); ok {
				err = nil
			} else {
				return
			}
		} else {
			available = true
		}
	}

	if !available {
		err = utils.Exec("", "cp", dskPth, tmpPath)
		if err != nil {
			return
		}
	}

	err = utils.Chmod(tmpPath, 0600)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"disk_id":    dsk.Id.Hex(),
		"disk_path":  dskPth,
		"storage_id": store.Id.Hex(),
		"object_key": img.Key,
	}).Info("data: Uploading disk backup")

	client, err := minio.New(store.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(store.AccessKey, store.SecretKey, ""),
		Secure: !store.Insecure,
	})
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "data: Failed to connect to storage"),
		}
		return
	}

	putOpts := minio.PutObjectOptions{}
	storageClass := storage.FormatStorageClass(dc.BackupStorageClass)
	if storageClass != "" {
		putOpts.StorageClass = storageClass
	}

	_, err = client.FPutObject(context.Background(),
		store.Bucket, img.Key, tmpPath, putOpts)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "data: Failed to write object"),
		}

		return
	}

	time.Sleep(3 * time.Second)

	obj, err := client.StatObject(context.Background(),
		store.Bucket, img.Key, minio.StatObjectOptions{})
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "data: Failed to stat object"),
		}
		return
	}

	img.Etag = image.GetEtag(obj)
	img.LastModified = obj.LastModified

	if store.IsOracle() {
		img.StorageClass = storage.ParseStorageClass(obj)
	} else {
		img.StorageClass = dc.BackupStorageClass
	}

	err = img.Upsert(db)
	if err != nil {
		return
	}

	event.PublishDispatch(db, "image.change")

	return
}

func RestoreBackup(db *database.Database, dsk *disk.Disk) (err error) {
	dskPth := paths.GetDiskPath(dsk.Id)
	cacheDir := node.Self.GetCachePath()

	img, err := image.Get(db, dsk.RestoreImage)
	if err != nil {
		return
	}

	if img.Disk != dsk.Id {
		err = &errortypes.VerificationError{
			errors.Wrap(err, "data: Restore image invalid"),
		}
		return
	}

	store, err := storage.Get(db, img.Storage)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"disk_id":    dsk.Id.Hex(),
		"image_id":   img.Id.Hex(),
		"storage_id": store.Id.Hex(),
		"disk_path":  dskPth,
	}).Info("data: Restoring disk backup")

	err = utils.ExistsMkdir(cacheDir, 0755)
	if err != nil {
		return
	}

	client, err := minio.New(store.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(store.AccessKey, store.SecretKey, ""),
		Secure: !store.Insecure,
	})
	if err != nil {
		err = &errortypes.ConnectionError{
			errors.Wrap(err, "data: Failed to connect to storage"),
		}
		return
	}

	imgId := primitive.NewObjectID()
	tmpPath := path.Join(cacheDir,
		fmt.Sprintf("restore-%s", imgId.Hex()))

	defer utils.Remove(tmpPath)
	err = client.FGetObject(context.Background(), store.Bucket,
		img.Key, tmpPath, minio.GetObjectOptions{})
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "data: Failed to download restore image"),
		}
		return
	}

	err = utils.Chmod(tmpPath, 0600)
	if err != nil {
		return
	}

	err = utils.Exec("", "mv", "-f", tmpPath, dskPth)
	if err != nil {
		return
	}

	return
}

func ImageAvailable(store *storage.Storage, img *image.Image) (
	available bool, err error) {

	if strings.Contains(strings.ToLower(store.Endpoint), "oracle") {
		client, e := minio.New(store.Endpoint, &minio.Options{
			Creds: credentials.NewStaticV4(store.AccessKey,
				store.SecretKey, ""),
			Secure: !store.Insecure,
		})
		if e != nil {
			err = &errortypes.ConnectionError{
				errors.Wrap(e, "data: Failed to connect to storage"),
			}
			return
		}

		obj, e := client.StatObject(context.Background(),
			store.Bucket, img.Key, minio.StatObjectOptions{})
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "data: Failed to stat object"),
			}
			return
		}

		archivalState := strings.ToLower(obj.Metadata.Get("Archival-State"))
		if archivalState != "" && archivalState != "restored" {
			available = false
			return
		}

		available = true
		return
	}

	switch img.StorageClass {
	case storage.AwsStandard:
		available = true
		break
	case storage.AwsInfrequentAccess:
		available = true
		break
	case storage.AwsGlacier:
		client, e := minio.New(store.Endpoint, &minio.Options{
			Creds: credentials.NewStaticV4(store.AccessKey,
				store.SecretKey, ""),
			Secure: !store.Insecure,
		})
		if e != nil {
			err = &errortypes.ConnectionError{
				errors.Wrap(e, "data: Failed to connect to storage"),
			}
			return
		}

		obj, e := client.StatObject(context.Background(),
			store.Bucket, img.Key, minio.StatObjectOptions{})
		if e != nil {
			err = &errortypes.ReadError{
				errors.Wrap(e, "data: Failed to stat object"),
			}
			return
		}

		restore := obj.Metadata.Get("x-amz-restore")
		if strings.Contains(restore, "ongoing-request=\"false\"") &&
			strings.Contains(restore, "expiry-date") {

			available = true
		} else {
			available = false
		}
		break
	default:
		available = true
		break
	}

	return
}
