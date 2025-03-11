package oracle

import (
	"encoding/json"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Metadata struct {
	UserOcid        string
	PrivateKey      string
	RegionName      string
	TenancyOcid     string
	CompartmentOcid string
	InstanceOcid    string
	VnicOcid        string
}

type OciMetaVnic struct {
	Id                  string   `json:"vnicId"`
	VlanTag             int      `json:"vlanTag"`
	MacAddr             string   `json:"macAddr"`
	PrivateIp           string   `json:"privateIp"`
	VirtualRouterIp     string   `json:"virtualRouterIp"`
	SubnetCidrBlock     string   `json:"subnetCidrBlock"`
	Ipv6Addresses       []string `json:"ipv6Addresses"`
	Ipv6SubnetCidrBlock string   `json:"ipv6SubnetCidrBlock"`
	Ipv6VirtualRouterIp string   `json:"ipv6VirtualRouterIp"`
	NicIndex            int      `json:"nicIndex"`
}

type OciMetaInstance struct {
	Id            string `json:"id"`
	DisplayName   string `json:"displayName"`
	CompartmentId string `json:"compartmentId"`
	RegionName    string `json:"canonicalRegionName"`
	Shape         string `json:"shape"`
}

type OciMeta struct {
	Instance OciMetaInstance `json:"instance"`
	Vnics    []OciMetaVnic   `json:"vnics"`
}

func (o *OciMeta) IsBareMetal() bool {
	if strings.Contains(o.Instance.Shape, "BM.") {
		return true
	}
	return false
}

func GetMetadata(authPv AuthProvider) (mdata *Metadata, err error) {
	userOcid := authPv.OracleUser()
	tenancyOcid := authPv.OracleTenancy()
	privateKey := authPv.OraclePrivateKey()

	output, err := utils.ExecOutput("", "oci-metadata", "--json")
	if err != nil {
		return
	}

	data := &OciMeta{}

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
		TenancyOcid:     tenancyOcid,
		CompartmentOcid: data.Instance.CompartmentId,
		InstanceOcid:    data.Instance.Id,
		VnicOcid:        vnicOcid,
	}

	return
}

func GetOciMetadata() (mdata *OciMeta, err error) {
	output, err := utils.ExecOutput("", "oci-metadata", "--json")
	if err != nil {
		return
	}

	mdata = &OciMeta{}

	err = json.Unmarshal([]byte(output), mdata)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "oracle: Failed to parse metadata"),
		}
		return
	}

	return
}
