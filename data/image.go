package data

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/disk"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/lock"
	"github.com/pritunl/pritunl-cloud/lvm"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/qmp"
	"github.com/pritunl/pritunl-cloud/settings"
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
	nbdLock          = sync.Mutex{}
)

func getImageS3(db *database.Database, store *storage.Storage,
	dsk *disk.Disk, img *image.Image) (tmpPth string, err error) {

	tmpPth = paths.GetImageTempPath()

	logrus.WithFields(logrus.Fields{
		"image_id":   img.Id.Hex(),
		"storage_id": store.Id.Hex(),
		"key":        img.Key,
		"temp_path":  tmpPth,
	}).Info("data: Downloading s3 image")

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

	stat, err := client.StatObject(context.Background(), store.Bucket,
		img.Key, minio.StatObjectOptions{})
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "data: Failed to stat s3 image"),
		}
		return
	}

	prog := NewProgressS3(db, dsk, img, tmpPth, stat.Size)
	prog.Start()
	defer prog.Stop()

	err = client.FGetObject(context.Background(), store.Bucket,
		img.Key, tmpPth, minio.GetObjectOptions{})
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "data: Failed to download s3 image"),
		}
		return
	}

	prog.Stop()

	logrus.WithFields(logrus.Fields{
		"image_id":   img.Id.Hex(),
		"storage_id": store.Id.Hex(),
		"key":        img.Key,
		"temp_path":  tmpPth,
	}).Info("data: Downloaded s3 image")

	return
}

type ProgressS3 struct {
	db         *database.Database
	disk       *disk.Disk
	img        *image.Image
	done       chan bool
	stopOnce   sync.Once
	baseDir    string
	outPrefix  string
	Total      int64
	Wrote      int64
	LastWrote  int64
	LastReport int
	LastTime   time.Time
}

func NewProgressS3(db *database.Database, dsk *disk.Disk, img *image.Image,
	outPath string, size int64) (prog *ProgressS3) {

	prog = &ProgressS3{
		db:        db,
		disk:      dsk,
		img:       img,
		done:      make(chan bool),
		baseDir:   filepath.Dir(outPath),
		outPrefix: filepath.Base(outPath),
		Total:     size,
		LastTime:  time.Now(),
	}

	return
}

func (p *ProgressS3) Start() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-p.done:
				return
			case <-ticker.C:
				p.calculateProgress()
				p.syncProgress()
			}
		}
	}()
}

func (p *ProgressS3) Stop() {
	p.stopOnce.Do(func() {
		p.done <- true
		close(p.done)
	})
}

func (p *ProgressS3) calculateProgress() {
	var totalBytes int64 = 0

	files, err := os.ReadDir(p.baseDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), p.outPrefix) {
			info, err := file.Info()
			if err != nil {
				continue
			}
			totalBytes += info.Size()
		}
	}

	p.Wrote = totalBytes
}

func (p *ProgressS3) syncProgress() {
	percent := int(float64(p.Wrote) / float64(p.Total) * 100)
	if percent > 100 {
		percent = 100
	}

	if percent >= p.LastReport+10 {
		now := time.Now()
		elapsed := now.Sub(p.LastTime).Seconds()

		speed := float64(p.Wrote-p.LastWrote) / elapsed

		p.LastTime = now
		p.LastWrote = p.Wrote
		p.LastReport = percent - (percent % 10)

		if p.disk != nil && !p.disk.Instance.IsZero() {
			_ = instance.SetDownloadProgress(
				p.db, p.disk.Instance, p.LastReport, speed/1_000_000.0)
		}
	}

	return
}

type Progress struct {
	db         *database.Database
	disk       *disk.Disk
	img        *image.Image
	Total      int64
	Wrote      int64
	LastWrote  int64
	LastReport int
	LastTime   time.Time
}

func humanReadableSpeed(bytesPerSecond float64) string {
	switch {
	case bytesPerSecond >= 1_000_000_000:
		return fmt.Sprintf("%.2f GB/s", bytesPerSecond/1_000_000_000)
	case bytesPerSecond >= 1_000_000:
		return fmt.Sprintf("%.2f MB/s", bytesPerSecond/1_000_000)
	case bytesPerSecond >= 1_000:
		return fmt.Sprintf("%.2f KB/s", bytesPerSecond/1_000)
	default:
		return fmt.Sprintf("%.2f B/s", bytesPerSecond)
	}
}

func (p *Progress) Write(data []byte) (n int, err error) {
	n = len(data)
	p.Wrote += int64(n)

	percent := int(float64(p.Wrote) / float64(p.Total) * 100)
	if percent >= p.LastReport+10 {
		now := time.Now()
		elapsed := now.Sub(p.LastTime).Seconds()

		speed := float64(p.Wrote-p.LastWrote) / elapsed

		p.LastTime = now
		p.LastWrote = p.Wrote
		p.LastReport = percent - (percent % 10)

		if p.disk != nil && !p.disk.Instance.IsZero() {
			_ = instance.SetDownloadProgress(
				p.db, p.disk.Instance, p.LastReport, speed/1_000_000.0)
		}
	}

	return
}

func getImageWeb(db *database.Database, store *storage.Storage,
	dsk *disk.Disk, img *image.Image) (tmpPth string, err error) {

	tmpPth = paths.GetImageTempPath()

	logrus.WithFields(logrus.Fields{
		"image_id":   img.Id.Hex(),
		"storage_id": store.Id.Hex(),
		"key":        img.Key,
		"temp_path":  tmpPth,
	}).Info("data: Downloading web image")

	u := store.GetWebUrl()
	u.Path += "/" + img.Key

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "data: Failed to create file request"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl-cloud")

	resp, err := clientLarge.Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "data: File request error"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = &errortypes.RequestError{
			errors.Newf(
				"data: Bad status %d from file request",
				resp.StatusCode,
			),
		}
		return
	}

	contentLen, err := strconv.ParseInt(
		resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "data: Invalid content length from file request"),
		}
		return
	}

	if contentLen <= 0 {
		err = &errortypes.RequestError{
			errors.Wrap(err, "data: Zero content length from file request"),
		}
		return
	}

	out, err := os.Create(tmpPth)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "data: Failed to create temporary file"),
		}
		return
	}
	defer out.Close()

	prog := &Progress{
		db:       db,
		disk:     dsk,
		img:      img,
		Total:    contentLen,
		LastTime: time.Now(),
	}

	_, err = io.Copy(out, io.TeeReader(resp.Body, prog))
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "data: Failed to download file"),
		}
		return
	}

	return
}

func checkImageSigS3(db *database.Database, store *storage.Storage,
	img *image.Image, tmpPth string) (err error) {

	sigPth := tmpPth + ".sig"
	defer os.Remove(sigPth)

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
		img.Key+".sig", sigPth, minio.GetObjectOptions{})
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "data: Failed to download image signature"),
		}
		return
	}

	signature, err := os.Open(sigPth)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "data: Failed to open image signature"),
		}
		return
	}
	defer signature.Close()

	tmpImg, err := os.Open(tmpPth)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "data: Failed to open image"),
		}
		return
	}
	defer tmpImg.Close()

	keyring, err := openpgp.ReadArmoredKeyRing(
		strings.NewReader(constants.PritunlKeyring))
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "data: Failed to parse Pritunl keyring"),
		}
		return
	}

	entity, err := openpgp.CheckArmoredDetachedSignature(
		keyring, tmpImg, signature)
	if err != nil || entity == nil {
		err = &errortypes.VerificationError{
			errors.Wrap(err, "data: Image signature verification failed"),
		}
		return
	}

	logrus.WithFields(logrus.Fields{
		"id":         img.Id.Hex(),
		"storage_id": store.Id.Hex(),
		"key":        img.Key,
	}).Info("data: Image signature successfully validated")

	return
}

func checkImageSigWeb(db *database.Database, store *storage.Storage,
	img *image.Image, tmpPth string) (err error) {

	sigPth := tmpPth + ".sig"
	defer os.Remove(sigPth)

	u := store.GetWebUrl()
	u.Path += "/" + img.Key + ".sig"

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "data: Failed to create file request"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl-cloud")

	resp, err := client.Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "data: File request error"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = &errortypes.RequestError{
			errors.Newf(
				"data: Bad status %d from file request",
				resp.StatusCode,
			),
		}
		return
	}

	tmpImg, err := os.Open(tmpPth)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "data: Failed to open image"),
		}
		return
	}
	defer tmpImg.Close()

	keyring, err := openpgp.ReadArmoredKeyRing(
		strings.NewReader(constants.PritunlKeyring))
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "data: Failed to parse Pritunl keyring"),
		}
		return
	}

	entity, err := openpgp.CheckArmoredDetachedSignature(
		keyring, tmpImg, resp.Body)
	if err != nil || entity == nil {
		err = &errortypes.VerificationError{
			errors.Wrap(err, "data: Image signature verification failed"),
		}
		return
	}

	logrus.WithFields(logrus.Fields{
		"id":         img.Id.Hex(),
		"storage_id": store.Id.Hex(),
		"key":        img.Key,
	}).Info("data: Image signature successfully validated")

	return
}

func getImage(db *database.Database, dsk *disk.Disk, img *image.Image,
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

	tmpPth := ""
	defer func() {
		if tmpPth != "" {
			utils.Remove(tmpPth)
		}
	}()

	exists, err := utils.Exists(pth)
	if err != nil {
		return
	}

	if exists {
		return
	}

	store, err := storage.Get(db, img.Storage)
	if err != nil {
		return
	}

	if img.Type == storage.Web {
		tmpPth, err = getImageWeb(db, store, dsk, img)
		if err != nil {
			return
		}
	} else {
		tmpPth, err = getImageS3(db, store, dsk, img)
		if err != nil {
			return
		}
	}

	if img.Signed || store.Endpoint == "images.pritunl.com" {
		if img.Type == storage.Web {
			err = checkImageSigWeb(db, store, img, tmpPth)
			if err != nil {
				return
			}
		} else {
			err = checkImageSigS3(db, store, img, tmpPth)
			if err != nil {
				return
			}
		}
	}

	hashed := false
	if img.Hash != "" {
		hash, e := utils.FileSha256(tmpPth)
		if e != nil {
			err = e
			return
		}

		if hash != img.Hash {
			err = &errortypes.VerificationError{
				errors.Wrap(err, "data: Image hash verification failed"),
			}
			return
		}

		hashed = true
	}

	logrus.WithFields(logrus.Fields{
		"image_id":   img.Id.Hex(),
		"storage_id": store.Id.Hex(),
		"key":        img.Key,
		"temp_path":  tmpPth,
		"path":       pth,
		"hashed":     hashed,
	}).Info("data: Downloaded image")

	err = utils.Exec("", "mv", tmpPth, pth)
	if err != nil {
		return
	}
	tmpPth = ""

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

func writeFsQcow(db *database.Database, dsk *disk.Disk) (err error) {
	ndbPath := settings.Hypervisor.NbdPath

	nbdLock.Lock()
	defer func() {
		utils.Exec("", "sync")
		utils.Exec("", "qemu-nbd", "--disconnect", ndbPath)
		nbdLock.Unlock()
	}()

	diskPath := paths.GetDiskPath(dsk.Id)

	err = utils.Exec("", "qemu-img", "create",
		"-f", "qcow2", diskPath, fmt.Sprintf("%dG", dsk.Size))
	if err != nil {
		return
	}

	err = utils.Chmod(diskPath, 0600)
	if err != nil {
		return
	}

	err = utils.Exec("", "modprobe", "nbd")
	if err != nil {
		return
	}

	err = utils.Exec("", "qemu-nbd", "--disconnect", ndbPath)
	if err != nil {
		return
	}

	err = utils.Exec("", "qemu-nbd", "--connect", ndbPath, diskPath)
	if err != nil {
		return
	}

	err = utils.Exec("", "parted", "--script", ndbPath, "mklabel", "gpt")
	if err != nil {
		return
	}

	err = utils.Exec("", "parted", "--script", ndbPath, "mkpart",
		"primary", "1MiB", "100%")
	if err != nil {
		return
	}

	time.Sleep(1 * time.Second)

	diskFs := ""
	diskLvm := false
	switch dsk.FileSystem {
	case disk.Xfs:
		diskFs = "xfs"
		diskLvm = false
	case disk.LvmXfs:
		diskFs = "xfs"
		diskLvm = true
	case disk.Ext4:
		diskFs = "ext4"
		diskLvm = false
	case disk.LvmExt4:
		diskFs = "ext4"
		diskLvm = true
	default:
		err = &errortypes.WriteError{
			errors.Newf("data: Invalid disk filesystem %s", dsk.FileSystem),
		}
		return
	}

	if diskLvm {
		vgName := GetVgName(dsk.Id, 0)
		lvName := GetLvName(dsk.Id, 0)

		err = utils.Exec("", "pvcreate", ndbPath+"p1")
		if err != nil {
			return
		}

		err = utils.Exec("", "vgcreate", vgName, ndbPath+"p1")
		if err != nil {
			return
		}

		if dsk.LvSize == dsk.Size {
			err = utils.Exec("", "lvcreate", "-l", "100%",
				"-n", lvName, vgName)
			if err != nil {
				return
			}
		} else {
			err = utils.Exec("", "lvcreate", "-L",
				fmt.Sprintf("%dG", dsk.LvSize), "-n", lvName, vgName)
			if err != nil {
				return
			}
		}

		err = utils.Exec("", "mkfs", "-t", diskFs,
			fmt.Sprintf("/dev/%s/%s", vgName, lvName))
		if err != nil {
			return
		}

		time.Sleep(100 * time.Millisecond)

		output, e := utils.ExecOutput("", "blkid", "-s", "UUID",
			"-o", "value", fmt.Sprintf("/dev/%s/%s", vgName, lvName))
		if e != nil {
			err = e
			return
		}

		dsk.Uuid = strings.TrimSpace(output)

		err = utils.Exec("", "lvchange", "-an",
			fmt.Sprintf("/dev/%s/%s", vgName, lvName))
		if err != nil {
			return
		}

		time.Sleep(50 * time.Millisecond)

		err = utils.Exec("", "vgchange", "-an", vgName)
		if err != nil {
			return
		}

		time.Sleep(50 * time.Millisecond)
	} else {
		err = utils.Exec("", "mkfs", "-t", diskFs, ndbPath+"p1")
		if err != nil {
			return
		}

		time.Sleep(100 * time.Millisecond)

		output, e := utils.ExecOutput("", "blkid", "-s", "UUID",
			"-o", "value", ndbPath+"p1")
		if e != nil {
			err = e
			return
		}

		dsk.Uuid = strings.TrimSpace(output)
	}

	err = dsk.CommitFields(db, set.NewSet("uuid"))
	if err != nil {
		return
	}

	return
}

func writeImageQcow(db *database.Database, dsk *disk.Disk) (
	newSize int, backingImageName string, err error) {

	size := dsk.Size
	diskPath := paths.GetDiskPath(dsk.Id)
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

	img, err := image.Get(db, dsk.Image)
	if err != nil {
		return
	}

	largeBase := strings.Contains(img.Key, "fedora")

	backingImagePth := path.Join(
		backingPath,
		fmt.Sprintf("image-%s-%s", img.Id.Hex(), img.Etag),
	)

	backingImageExists := false
	if dsk.Backing {
		backingImageName = fmt.Sprintf("%s-%s", img.Id.Hex(), img.Etag)

		backingImageExists, err = utils.Exists(backingImagePth)
		if err != nil {
			return
		}
	}

	if img.Type == storage.Public || img.Type == storage.Web ||
		!img.Deployment.IsZero() {

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
			err = getImage(db, dsk, img, imagePth)
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
				"disk_id":    dsk.Id.Hex(),
				"key":        img.Key,
				"path":       diskPath,
			}).Error("data: Blocking disk image overwrite")

			err = &errortypes.WriteError{
				errors.Wrap(err, "data: Image already exists"),
			}
			return
		}

		utils.Exec("", "touch", imagePth)

		if dsk.Backing {
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
		if dsk.Backing {
			err = getImage(db, dsk, img, backingImagePth)
			if err != nil {
				return
			}
		} else {
			err = getImage(db, dsk, img, diskTempPath)
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
				"disk_id":    dsk.Id.Hex(),
				"key":        img.Key,
				"path":       diskPath,
			}).Error("data: Blocking disk image overwrite")

			err = &errortypes.WriteError{
				errors.Wrap(err, "data: Image already exists"),
			}
			return
		}

		if dsk.Backing {
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

func writeFsLvm(db *database.Database, dsk *disk.Disk,
	pl *pool.Pool) (err error) {

	size := dsk.Size
	vgName := pl.VgName
	lvName := dsk.Id.Hex()
	sourcePth := ""
	diskTempPath := paths.GetDiskTempPath()
	defer utils.Remove(diskTempPath)

	acquired, err := lock.LvmLock(db, vgName, lvName)
	if err != nil {
		return
	}

	if !acquired {
		err = &errortypes.WriteError{
			errors.New("data: Failed to acquire LVM lock"),
		}
		return
	}
	defer func() {
		err2 := lock.LvmUnlock(db, vgName, lvName)
		if err2 != nil {
			logrus.WithFields(logrus.Fields{
				"error": err2,
			}).Error("data: Failed to unlock lvm")
		}
	}()

	diskFs := ""
	diskLvm := false
	switch dsk.FileSystem {
	case disk.Xfs:
		diskFs = "xfs"
		diskLvm = false
	case disk.LvmXfs:
		diskFs = "xfs"
		diskLvm = true
	case disk.Ext4:
		diskFs = "ext4"
		diskLvm = false
	case disk.LvmExt4:
		diskFs = "ext4"
		diskLvm = true
	default:
		err = &errortypes.WriteError{
			errors.Newf("data: Invalid disk filesystem %s", dsk.FileSystem),
		}
		return
	}

	err = lvm.CreateLv(vgName, lvName, size)
	if err != nil {
		return
	}

	err = lvm.ActivateLv(vgName, lvName)
	if err != nil {
		return
	}

	defer func() {
		err = lvm.DeactivateLv(vgName, lvName)
		if err != nil {
			return
		}
	}()

	err = lvm.WriteLv(vgName, lvName, sourcePth)
	if err != nil {
		return
	}

	diskPath := filepath.Join("/dev/mapper",
		fmt.Sprintf("%s-%s", vgName, lvName))

	if diskLvm {
		vgName := GetVgName(dsk.Id, 0)
		lvName := GetLvName(dsk.Id, 0)

		err = utils.Exec("", "pvcreate", diskPath)
		if err != nil {
			return
		}

		err = utils.Exec("", "vgcreate", vgName, diskPath)
		if err != nil {
			return
		}

		if dsk.LvSize == dsk.Size {
			err = utils.Exec("", "lvcreate", "-l", "100%",
				"-n", lvName, vgName)
			if err != nil {
				return
			}
		} else {
			err = utils.Exec("", "lvcreate", "-L",
				fmt.Sprintf("%dG", dsk.LvSize), "-n", lvName, vgName)
			if err != nil {
				return
			}
		}

		err = utils.Exec("", "mkfs", "-t", diskFs,
			fmt.Sprintf("/dev/%s/%s", vgName, lvName))
		if err != nil {
			return
		}

		time.Sleep(100 * time.Millisecond)

		output, e := utils.ExecOutput("", "blkid", "-s", "UUID",
			"-o", "value", fmt.Sprintf("/dev/%s/%s", vgName, lvName))
		if e != nil {
			err = e
			return
		}

		dsk.Uuid = strings.TrimSpace(output)

		err = utils.Exec("", "lvchange", "-an",
			fmt.Sprintf("/dev/%s/%s", vgName, lvName))
		if err != nil {
			return
		}

		time.Sleep(50 * time.Millisecond)

		err = utils.Exec("", "vgchange", "-an", vgName)
		if err != nil {
			return
		}

		time.Sleep(50 * time.Millisecond)
	} else {
		err = utils.Exec("", "mkfs", "-t", diskFs, diskPath)
		if err != nil {
			return
		}

		output, e := utils.ExecOutput("", "blkid", "-s", "UUID",
			"-o", "value", diskPath)
		if e != nil {
			err = e
			return
		}

		dsk.Uuid = strings.TrimSpace(output)
	}

	err = dsk.CommitFields(db, set.NewSet("uuid"))
	if err != nil {
		return
	}

	return
}

func writeImageLvm(db *database.Database, dsk *disk.Disk,
	pl *pool.Pool) (newSize int, err error) {

	size := dsk.Size
	vgName := pl.VgName
	lvName := dsk.Id.Hex()
	sourcePth := ""
	diskTempPath := paths.GetDiskTempPath()
	defer utils.Remove(diskTempPath)

	if dsk.Backing {
		err = &errortypes.ParseError{
			errors.New("data: Cannot create LVM disk with linked image"),
		}
		return
	}

	img, err := image.Get(db, dsk.Image)
	if err != nil {
		return
	}

	largeBase := strings.Contains(img.Key, "fedora")

	if img.Type == storage.Public || img.Type == storage.Web ||
		!img.Deployment.IsZero() {

		cacheDir := node.Self.GetCachePath()

		imagePth := path.Join(
			cacheDir,
			fmt.Sprintf("image-%s-%s", img.Id.Hex(), img.Etag),
		)

		err = utils.ExistsMkdir(cacheDir, 0755)
		if err != nil {
			return
		}

		err = getImage(db, dsk, img, imagePth)
		if err != nil {
			return
		}

		sourcePth = imagePth

		if largeBase && size < 16 {
			size = 16
			newSize = 16
		} else if !largeBase && size < 10 {
			size = 10
			newSize = 10
		}
	} else {
		err = getImage(db, dsk, img, diskTempPath)
		if err != nil {
			return
		}

		sourcePth = diskTempPath

		if largeBase && size < 16 {
			size = 16
			newSize = 16
		}
	}

	acquired, err := lock.LvmLock(db, vgName, lvName)
	if err != nil {
		return
	}

	if !acquired {
		err = &errortypes.WriteError{
			errors.New("data: Failed to acquire LVM lock"),
		}
		return
	}
	defer func() {
		err2 := lock.LvmUnlock(db, vgName, lvName)
		if err2 != nil {
			logrus.WithFields(logrus.Fields{
				"error": err2,
			}).Error("data: Failed to unlock lvm")
		}
	}()

	err = lvm.CreateLv(vgName, lvName, size)
	if err != nil {
		return
	}

	err = lvm.ActivateLv(vgName, lvName)
	if err != nil {
		return
	}

	defer func() {
		err = lvm.DeactivateLv(vgName, lvName)
		if err != nil {
			return
		}
	}()

	err = lvm.WriteLv(vgName, lvName, sourcePth)
	if err != nil {
		return
	}

	return
}

func DeleteImage(db *database.Database, imgId primitive.ObjectID) (
	err error) {

	img, err := image.Get(db, imgId)
	if err != nil {
		return
	}

	if img.Type == storage.Public || img.Type == storage.Web {
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

	err = img.Remove(db)
	if err != nil {
		return
	}

	return
}

func WriteImage(db *database.Database, dsk *disk.Disk) (
	newSize int, backingImageName string, err error) {

	switch dsk.Type {
	case disk.Lvm:
		pl, e := pool.Get(db, dsk.Pool)
		if e != nil {
			err = e
			return
		}

		err = lvm.InitLock(pl.VgName)
		if err != nil {
			return
		}

		newSize, err = writeImageLvm(db, dsk, pl)
		if err != nil {
			return
		}
		break
	case "", disk.Qcow2:
		newSize, backingImageName, err = writeImageQcow(db, dsk)
		if err != nil {
			return
		}
		break
	default:
		err = &errortypes.ParseError{
			errors.Newf("data: Unknown disk type %s", dsk.Type),
		}
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

	if img.Type == storage.Public || img.Type == storage.Web {
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

	err = img.Remove(db)
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

	if store.Type != storage.Private {
		err = &errortypes.ConnectionError{
			errors.New("data: Cannot upload to non-private storage"),
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
			time.Now().Format("20060102-150405")),
		Organization: dsk.Organization,
		Deployment:   dsk.Deployment,
		Type:         storage.Private,
		SystemType:   dsk.SystemType,
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

	hash, err := utils.FileSha256(tmpPath)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"disk_id":    dsk.Id.Hex(),
		"disk_path":  dskPth,
		"storage_id": store.Id.Hex(),
		"object_key": img.Key,
		"hash":       hash,
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

	img.Hash = hash
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

	if !dsk.Deployment.IsZero() {
		deply, e := deployment.Get(db, dsk.Deployment)
		if e != nil {
			err = e
			return
		}

		deply.Image = img.Id
		deply.SetImageState(deployment.Complete)
		err = deply.CommitFields(db, set.NewSet(
			"image", "image_data.state"))
		if err != nil {
			return
		}

		err = instance.Delete(db, deply.Instance)
		if err != nil {
			return
		}
	}

	logrus.WithFields(logrus.Fields{
		"disk_id":    dsk.Id.Hex(),
		"disk_path":  dskPth,
		"storage_id": store.Id.Hex(),
		"object_key": img.Key,
	}).Info("data: Uploaded disk snapshot")

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

	if store.Type != storage.Private {
		err = &errortypes.ConnectionError{
			errors.New("data: Cannot upload to non-private storage"),
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
			time.Now().Format("20060102-150405")),
		Organization: dsk.Organization,
		Type:         storage.Private,
		SystemType:   dsk.SystemType,
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

	hash, err := utils.FileSha256(tmpPath)
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"disk_id":    dsk.Id.Hex(),
		"disk_path":  dskPth,
		"storage_id": store.Id.Hex(),
		"object_key": img.Key,
		"hash":       hash,
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

	img.Hash = hash
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

	if store.Type != storage.Private {
		err = &errortypes.ConnectionError{
			errors.New("data: Cannot restore from non-private storage"),
		}
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

	hashed := false
	if img.Hash != "" {
		hash, e := utils.FileSha256(tmpPath)
		if e != nil {
			err = e
			return
		}

		if hash != img.Hash {
			err = &errortypes.VerificationError{
				errors.Wrap(err, "data: Image hash verification failed"),
			}
			return
		}

		hashed = true
	}

	logrus.WithFields(logrus.Fields{
		"image_id":   img.Id.Hex(),
		"storage_id": store.Id.Hex(),
		"key":        img.Key,
		"temp_path":  tmpPath,
		"disk_path":  dskPth,
		"hashed":     hashed,
	}).Info("data: Restored backup")

	err = utils.Exec("", "mv", "-f", tmpPath, dskPth)
	if err != nil {
		return
	}

	return
}

func ImageAvailable(store *storage.Storage, img *image.Image) (
	available bool, err error) {

	if img.Type == storage.Web {
		available = true
		return
	}

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
