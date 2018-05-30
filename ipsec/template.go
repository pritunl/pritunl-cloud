package ipsec

import (
	"bytes"
	"fmt"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/link"
	"github.com/pritunl/pritunl-cloud/vm"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"path"
	"strings"
)

type templateData struct {
	Id           string
	Left         string
	LeftSubnets  string
	Right        string
	RightSubnets string
	PreSharedKey string
}

func writeTemplates(vpcId bson.ObjectId, states []*link.State) (err error) {
	namespace := vm.GetLinkNamespace(vpcId, 0)
	baseDir := path.Join("/", "etc", "netns", namespace)

	confBuf := &bytes.Buffer{}
	secretsBuf := &bytes.Buffer{}

	for _, stat := range states {
		for i, lnk := range stat.Links {
			leftSubnets := strings.Join(lnk.LeftSubnets, ",")
			rightSubnets := strings.Join(lnk.RightSubnets, ",")

			left := ""
			if stat.Ipv6 {
				left = stat.PublicAddr6
			} else {
				left = stat.PublicAddr
			}

			data := &templateData{
				Id:           fmt.Sprintf("%s-%d", stat.Id, i),
				Left:         left,
				LeftSubnets:  leftSubnets,
				Right:        lnk.Right,
				RightSubnets: rightSubnets,
				PreSharedKey: lnk.PreSharedKey,
			}

			err = confTemplate.Execute(confBuf, data)
			if err != nil {
				err = &errortypes.ParseError{
					errors.Wrap(err,
						"ipsec: Failed to execute conf template"),
				}
				return
			}

			err = secretsTemplate.Execute(secretsBuf, data)
			if err != nil {
				err = &errortypes.ParseError{
					errors.Wrap(err,
						"ipsec: Failed to execute secrets template"),
				}
				return
			}
		}
	}

	pth := path.Join(baseDir, "ipsec.conf")
	err = ioutil.WriteFile(pth, confBuf.Bytes(), 0644)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "ipsec: Failed to write state conf"),
		}
		return
	}

	pth = path.Join(baseDir, "ipsec.secrets")
	err = ioutil.WriteFile(pth, secretsBuf.Bytes(), 0600)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrap(err, "ipsec: Failed to write state secrets"),
		}
		return
	}

	return
}
