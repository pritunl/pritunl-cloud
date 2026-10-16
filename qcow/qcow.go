package qcow

import (
	"encoding/json"
	"fmt"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

type Info struct {
	Filename              string            `json:"filename"`
	Format                string            `json:"format"`
	VirtualSize           int64             `json:"virtual-size"`
	ActualSize            int64             `json:"actual-size"`
	ClusterSize           int64             `json:"cluster-size"`
	BackingFilename       string            `json:"backing-filename"`
	FullBackingFilename   string            `json:"full-backing-filename"`
	BackingFilenameFormat string            `json:"backing-filename-format"`
	Snapshots             []json.RawMessage `json:"snapshots"`
	FormatSpecific        *FormatSpecific   `json:"format-specific"`
}

type FormatSpecific struct {
	Type string     `json:"type"`
	Data FormatData `json:"data"`
}

type FormatData struct {
	Compat      string          `json:"compat"`
	DataFile    string          `json:"data-file"`
	DataFileRaw *bool           `json:"data-file-raw"`
	Encrypt     json.RawMessage `json:"encrypt"`
	Corrupt     bool            `json:"corrupt"`
}

func GetInfo(pth string) (info *Info, err error) {
	output, err := utils.ExecOutputLogged(nil, "qemu-img", "info",
		"-f", Format, "--output=json", pth)
	if err != nil {
		return
	}

	info = &Info{}
	err = json.Unmarshal([]byte(output), info)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "qcow: Failed to parse qemu image info"),
		}
		return
	}

	return
}

func (i *Info) check() (reason string) {
	if i.Format != Format {
		return fmt.Sprintf("format %s is not %s", i.Format, Format)
	}

	if i.BackingFilename != "" || i.FullBackingFilename != "" ||
		i.BackingFilenameFormat != "" {

		return "image has a backing file"
	}

	if i.FormatSpecific == nil || i.FormatSpecific.Type != Format {
		return "image missing qcow2 header details"
	}

	data := i.FormatSpecific.Data

	if data.DataFile != "" || data.DataFileRaw != nil {
		return "image has an external data file"
	}

	if len(data.Encrypt) != 0 && string(data.Encrypt) != "null" {
		return "image is encrypted"
	}

	if data.Corrupt {
		return "image is marked corrupt"
	}

	if i.VirtualSize <= 0 {
		return "image has no virtual size"
	}

	return ""
}

func Check(pth string) (info *Info, err error) {
	info, err = GetInfo(pth)
	if err != nil {
		return
	}

	reason := info.check()
	if reason != "" {
		logrus.WithFields(logrus.Fields{
			"path":   pth,
			"reason": reason,
		}).Error("qcow: Rejected image")

		err = &errortypes.VerificationError{
			errors.Newf("qcow: Rejected image %s, %s", pth, reason),
		}
		return
	}

	return
}

func Source(pth string) (src string, err error) {
	opts := map[string]interface{}{
		"driver": Format,
		"file": map[string]interface{}{
			"driver":   "file",
			"filename": pth,
		},
		"backing": nil,
	}

	data, err := json.Marshal(opts)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "qcow: Failed to marshal image source"),
		}
		return
	}

	src = "json:" + string(data)

	return
}

func convert(srcPth, dstPth, format string) (err error) {
	_, err = Check(srcPth)
	if err != nil {
		return
	}

	src, err := Source(srcPth)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(nil, "qemu-img", "convert",
		"-O", format, src, dstPth)
	if err != nil {
		return
	}

	return
}

func Convert(srcPth, dstPth string) (err error) {
	err = convert(srcPth, dstPth, Format)
	if err != nil {
		return
	}

	err = utils.Chmod(dstPth, 0600)
	if err != nil {
		return
	}

	return
}

func ConvertRaw(srcPth, dstPth string) (err error) {
	err = convert(srcPth, dstPth, Raw)
	if err != nil {
		return
	}

	return
}

func Flatten(srcPth, dstPth string) (err error) {
	_, err = utils.ExecCombinedOutputLogged(nil, "qemu-img", "convert",
		"-f", Format, "-O", Format, srcPth, dstPth)
	if err != nil {
		return
	}

	err = utils.Chmod(dstPth, 0600)
	if err != nil {
		return
	}

	return
}

func Size(pth string) (sizeGb int, err error) {
	info, err := GetInfo(pth)
	if err != nil {
		return
	}

	sizeGb = int(info.VirtualSize / gib)

	return
}

func Resize(pth string, sizeGb int) (err error) {
	_, err = utils.ExecCombinedOutputLogged(nil, "qemu-img", "resize",
		"-f", Format, pth, fmt.Sprintf("%dG", sizeGb))
	if err != nil {
		return
	}

	return
}

func Expand(pth string, addGb int) (err error) {
	_, err = utils.ExecCombinedOutputLogged(nil, "qemu-img", "resize",
		"-f", Format, pth, fmt.Sprintf("+%dG", addGb))
	if err != nil {
		return
	}

	return
}
