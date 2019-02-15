package oracle

import (
	"github.com/pritunl/pritunl-cloud/state"
)

var (
	curState = ""
)

func ApplyState(stat *state.State) (err error) {
	nodeSelf := stat.Node()

	if !nodeSelf.OracleHostRoute || nodeSelf.OracleUser == "" {
		return
	}

	blck := stat.NodeHostBlock()
	if blck == nil {
		return
	}

	blckNet, err := blck.GetNetwork()
	if err != nil {
		return
	}

	if blckNet.String() == curState {
		return
	}

	mdata, err := GetMetadata(nodeSelf)
	if err != nil {
		return
	}

	pv, err := NewProvider(nodeSelf, mdata)
	if err != nil {
		return
	}

	vnic, err := GetVnic(pv, mdata.VnicOcid)
	if err != nil {
		return
	}

	if !vnic.SkipSourceDestCheck {
		err = vnic.SetSkipSourceDestCheck(pv, true)
		if err != nil {
			return
		}
	}

	subnet, err := GetSubnet(pv, vnic.SubnetId)
	if err != nil {
		return
	}

	tables, err := GetRouteTables(pv, subnet.VcnId)
	if err != nil {
		return
	}

	for _, table := range tables {
		if table.RouteUpsert(blckNet.String(), vnic.PrivateIpId) {
			err = table.CommitRouteRules(pv)
			if err != nil {
				return
			}
		}
	}

	curState = blckNet.String()

	return
}
