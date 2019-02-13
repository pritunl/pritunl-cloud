package oracle

import (
	"encoding/json"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Metadata struct {
	UserOcid        string
	PrivateKey      string
	RegionName      string
	TenancyOcid     string
	CompartmentOcid string
	VnicOcid        string
}

type ociMetaVnic struct {
	Id        string `json:"vnicId"`
	MacAddr   string `json:"macAddr"`
	PrivateIp string `json:"privateIp"`
}

type ociMetaInstance struct {
	Id            string `json:"id"`
	DisplayName   string `json:"displayName"`
	CompartmentId string `json:"compartmentId"`
	RegionName    string `json:"canonicalRegionName"`
}

type ociMeta struct {
	Instance ociMetaInstance `json:"instance"`
	Vnics    []ociMetaVnic   `json:"vnics"`
}

func GetMetadata() (mdata *Metadata, err error) {
	userOcid := node.Self.OracleUser
	privateKey := node.Self.OraclePrivateKey

	output, err := utils.ExecOutput("", "oci-metadata", "--json")
	if err != nil {
		return
	}

	data := &ociMeta{}

	err = json.Unmarshal([]byte(output), data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "oracle: Failed to parse metadata"),
		}
		return
	}

	vnicOcid := ""
	if data.Vnics != nil {
		for _, vnic := range data.Vnics {
			vnicOcid = vnic.Id
			break
		}
	}

	if vnicOcid == "" {
		err = &errortypes.ParseError{
			errors.Wrap(err, "oracle: Failed to get vnic in metadata"),
		}
		return
	}

	mdata = &Metadata{
		UserOcid:        userOcid,
		PrivateKey:      privateKey,
		RegionName:      data.Instance.RegionName,
		TenancyOcid:     data.Instance.CompartmentId,
		CompartmentOcid: data.Instance.CompartmentId,
		VnicOcid:        vnicOcid,
	}

	return
}
