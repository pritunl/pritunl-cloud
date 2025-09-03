package cloudinit

import (
	"encoding/json"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/tools/commander"
	"github.com/sirupsen/logrus"
)

type NetworkConfig struct {
	Config  []NetworkInterface `json:"config"`
	Version int                `json:"version"`
}

type NetworkInterface struct {
	Name           string   `json:"name"`
	Mtu            int      `json:"mtu,omitempty"`
	Type           string   `json:"type"`
	BondInterfaces []string `json:"bond_interfaces,omitempty"`
	Subnets        []Subnet `json:"subnets,omitempty"`
	VlanId         int      `json:"vlan_id,omitempty"`
	VlanLink       string   `json:"vlan_link,omitempty"`
}

type Subnet struct {
	Address string `json:"address"`
	Gateway string `json:"gateway,omitempty"`
	Type    string `json:"type"`
}

type CloudConfig struct {
	CombinedCloudConfig CombinedCloudConfig `json:"combined_cloud_config"`
	MergedSystemConfig  MergedSystemConfig  `json:"merged_system_cfg"`
}

type CombinedCloudConfig struct {
	Network NetworkConfig `json:"network"`
}

type MergedSystemConfig struct {
	Network NetworkConfig `json:"network"`
}

func GetCloudConfig() (data *CloudConfig, err error) {
	ret, err := commander.Exec(&commander.Opt{
		Name: "cloud-init",
		Args: []string{
			"query",
			"--all",
		},
		Timeout: 10 * time.Second,
		PipeOut: true,
		PipeErr: true,
	})
	if err != nil {
		if ret != nil {
			logrus.WithFields(ret.Map()).Warn(
				"cloudinit: Cloud init query failed")
		}
		return
	}

	data = &CloudConfig{}
	err = json.Unmarshal(ret.Output, &data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrapf(err, "cloudinit: Failed to parse cloudinit query"),
		}
		return
	}

	return
}
