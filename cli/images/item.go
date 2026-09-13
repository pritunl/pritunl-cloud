package images

import (
	"fmt"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"

	"github.com/pritunl/pritunl-cloud/cli/resource"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/image"
)

// distroRe splits a signed image release such as ubuntu2404 into the
// distribution and version like the web image rows.
var distroRe = regexp.MustCompile(`^(.+?)(\d+)$`)

// Item is an image card with the organization and storage names
// resolved.
type Item struct {
	img     *image.Image
	org     string
	storage string
}

func (i *Item) Image() *image.Image {
	return i.img
}

func (i *Item) Id() string {
	return i.img.Id.Hex()
}

// buildDate formats the four digit build of a signed image key as
// month/year like the web interface.
func buildDate(build string) string {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, build)

	if len(digits) < 4 {
		return build
	}
	year := digits[0:2]
	month := digits[2:4]
	if month < "01" || month > "12" {
		return build
	}

	return month + "/" + year
}

// signedName returns the display name of a signed public image from its
// key such as Ubuntu 24.04 (09/24), ok is false for unknown
// distributions.
func signedName(key string) (name string, ok bool) {
	parts := strings.Split(key, "_")
	build := ""
	if len(parts) > 1 {
		build = strings.SplitN(parts[1], ".", 2)[0]
	}

	distro := parts[0]
	version := ""
	if match := distroRe.FindStringSubmatch(parts[0]); match != nil {
		distro = match[1]
		version = match[2]
	}

	switch distro {
	case "almalinux":
		name = "AlmaLinux " + version
	case "alpinelinux":
		name = "Alpine Linux"
	case "archlinux":
		name = "Arch Linux"
	case "fedora":
		name = "Fedora " + version
	case "freebsd":
		name = "FreeBSD"
	case "oraclelinux":
		name = "Oracle Linux " + version
	case "rockylinux":
		name = "Rocky Linux " + version
	case "ubuntu":
		if len(version) > 2 {
			version = version[:2] + "." + version[2:]
		}
		name = "Ubuntu " + version
	default:
		return "", false
	}

	if build != "" {
		name += " (" + buildDate(build) + ")"
	}

	return name, true
}

// Name returns the image name, signed public images use the pretty
// distribution name of the web image rows.
func (i *Item) Name() string {
	if i.img.Signed {
		if name, ok := signedName(i.img.Key); ok {
			return name
		}
	}
	return i.img.Name
}

func (i *Item) Tag() string {
	return i.img.Id.Hex()
}

// organization returns the organization name or the public image
// labels of the web image rows.
func (i *Item) organization() string {
	if i.img.Signed {
		return "Signed Public Image"
	}
	if i.img.Organization.IsZero() {
		return "Public Image"
	}
	return widget.Default(i.org, i.img.Organization.Hex())
}

// storageClassLabel returns the display name of the storage class.
func storageClassLabel(class string) string {
	switch class {
	case "aws_standard":
		return "AWS Standard"
	case "aws_infrequent_access":
		return "AWS Standard-IA"
	case "aws_glacier":
		return "AWS Glacier"
	case "oracle_standard":
		return "Oracle Standard"
	case "oracle_archive":
		return "Oracle Archive"
	}
	return "Default"
}

func (i *Item) Fields() []resource.Field {
	return []resource.Field{
		{
			Label: "Organization",
			Value: i.organization(),
		},
		{
			Label: "Key",
			Value: i.img.Key,
		},
		{
			Label: "Storage Class",
			Value: storageClassLabel(i.img.StorageClass),
		},
		{
			Label: "Last Modified",
			Value: resource.FormatTime(i.img.LastModified),
		},
	}
}

// Info mirrors the fields of the detailed image view.
func (i *Item) Info() []widget.InfoField {
	img := i.img

	typ := img.Type
	if typ != "" {
		typ = strings.ToUpper(typ[:1]) + typ[1:]
	}

	fields := []widget.InfoField{
		{"ID", img.Id.Hex()},
		{"Storage", widget.Default(i.storage,
			widget.Default(resource.IdHex(img.Storage), "Unknown"))},
		{"Organization", i.organization()},
		{"Type", widget.Default(typ, "Unknown")},
		{"Key", widget.Default(img.Key, "Unknown")},
		{"Storage Class", storageClassLabel(img.StorageClass)},
		{"Last Modified", widget.Default(
			resource.FormatTime(img.LastModified), "Unknown")},
		{"ETag", widget.Default(img.Etag, "Unknown")},
	}

	if img.SystemType != "" || img.SystemKind != "" {
		system := img.SystemType
		if img.SystemType != "" && img.SystemKind != "" {
			system = fmt.Sprintf("%s/%s", img.SystemType, img.SystemKind)
		} else if system == "" {
			system = img.SystemKind
		}
		fields = append(fields, widget.InfoField{"System Type", system})
	}

	return fields
}

// deleteImage removes the image from its storage the same way as the
// admin image handler, pods are notified for their image lists.
func deleteImage(imgId bson.ObjectID) func(db *database.Database) error {
	return func(db *database.Database) (err error) {
		err = data.DeleteImage(db, imgId)
		if err != nil {
			return
		}

		logrus.WithFields(logrus.Fields{
			"image_id": imgId.Hex(),
		}).Info("tui: image deleted")

		err = event.PublishDispatch(db, "image.change")
		if err != nil {
			return
		}

		err = event.PublishDispatch(db, "pod.change")
		if err != nil {
			return
		}

		return
	}
}

// Actions mirrors the delete button of the detailed image view, public
// images cannot be deleted.
func (i *Item) Actions() []resource.Action {
	if i.img.Type == storage.Public {
		return nil
	}

	return []resource.Action{
		resource.DeleteAction("image", i.Name(), deleteImage(i.img.Id)),
	}
}

func (i *Item) Editor() resource.Editor {
	return newEditor(i.img)
}
