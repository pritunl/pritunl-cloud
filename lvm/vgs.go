package lvm

import (
	"encoding/json"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/utils"
)

type report struct {
	Report []*vgReport `json:"report"`
}

type vgReport struct {
	Vg []*vgDetails `json:"vg"`
}

type vgDetails struct {
	VgName    string `json:"vg_name"`
	PvCount   string `json:"pv_count"`
	LvCount   string `json:"lv_count"`
	SnapCount string `json:"snap_count"`
	VgAttr    string `json:"vg_attr"`
	VgSize    string `json:"vg_size"`
	VgFree    string `json:"vg_free"`
}

type VolumeGroup struct {
	Name string
}

func GetVgs() (vgs []*VolumeGroup, err error) {
	vgs = []*VolumeGroup{}

	output, err := utils.ExecCombinedOutput("",
		"vgs", "--reportformat", "json")
	if err != nil {
		return
	}

	reprt := &report{}
	err = json.Unmarshal([]byte(output), reprt)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "deploy: Failed to unmarshal vgs report"),
		}
		return
	}

	if reprt.Report == nil {
		return
	}

	for _, reportGroup := range reprt.Report {
		if reportGroup.Vg != nil {
			for _, reportVg := range reportGroup.Vg {
				vg := &VolumeGroup{
					Name: reportVg.VgName,
				}
				vgs = append(vgs, vg)
			}
		}
	}

	return
}

func GetVgsNameSet() (vgNames set.Set, err error) {
	vgNames = set.NewSet()

	output, err := utils.ExecCombinedOutput("",
		"vgs", "--reportformat", "json")
	if err != nil {
		return
	}

	reprt := &report{}
	err = json.Unmarshal([]byte(output), reprt)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "deploy: Failed to unmarshal vgs report"),
		}
		return
	}

	if reprt.Report == nil {
		return
	}

	for _, reportGroup := range reprt.Report {
		if reportGroup.Vg != nil {
			for _, reportVg := range reportGroup.Vg {
				vgNames.Add(reportVg.VgName)
			}
		}
	}

	return
}
