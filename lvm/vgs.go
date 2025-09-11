package lvm

import (
	"encoding/json"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	cachedNodePools          []*pool.Pool
	cachedNodePoolsTimestamp time.Time
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

func GetAvailablePools(db *database.Database, zoneId bson.ObjectID) (
	availablePools []*pool.Pool, err error) {

	if time.Since(cachedNodePoolsTimestamp) < 30*time.Second {
		availablePools = cachedNodePools
		return
	}

	availablePools = []*pool.Pool{}
	vgNames := set.NewSet()

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

	if reprt.Report != nil {
		for _, reportGroup := range reprt.Report {
			if reportGroup.Vg != nil {
				for _, reportVg := range reportGroup.Vg {
					vgNames.Add(reportVg.VgName)
				}
			}
		}
	}

	if vgNames.Len() > 0 {
		pools, e := pool.GetAll(db, &bson.M{
			"zone": zoneId,
		})
		if e != nil {
			err = e
			return
		}

		for _, pl := range pools {
			if vgNames.Contains(pl.VgName) {
				availablePools = append(availablePools, pl)
			}
		}
	}

	cachedNodePools = availablePools
	cachedNodePoolsTimestamp = time.Now()

	return
}
