package render

import (
	"io/ioutil"
	"path"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	renders         = []string{}
	lastRendersSync time.Time
)

func GetRenders() (rendrs []string, err error) {
	if time.Since(lastRendersSync) < 300*time.Second {
		rendrs = renders
		return
	}

	rendersNew := []string{}

	exists, err := utils.ExistsDir(RendersDir)
	if err != nil {
		return
	}

	if !exists {
		return
	}

	renderFiles, err := ioutil.ReadDir(RendersDir)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "backup: Failed to read renders directory"),
		}
		return
	}

	for _, item := range renderFiles {
		name := item.Name()
		if !strings.Contains(name, "render") {
			continue
		}

		rendersNew = append(rendersNew, item.Name())
	}

	renders = rendersNew
	lastRendersSync = time.Now()
	rendrs = rendersNew

	return
}

func GetRender(render string) (pth string, err error) {
	rendrs, err := GetRenders()
	if err != nil {
		return
	}

	for _, rendr := range rendrs {
		if rendr == render {
			pth = path.Join(RendersDir, rendr)
			return
		}
	}

	err = &errortypes.ReadError{
		errors.Newf("render: Failed to find render '%s'", render),
	}
	return
}
