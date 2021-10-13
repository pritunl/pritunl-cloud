package deploy

import (
	"github.com/pritunl/pritunl-cloud/oracle"
	"github.com/pritunl/pritunl-cloud/state"
)

var (
	curOracleState = ""
)

func ApplyOracleState(stat *state.State) (err error) {
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

	if blckNet.String() == curOracleState {
		return
	}

	pv, err := oracle.NewProvider(nodeSelf.GetOracleAuthProvider())
	if err != nil {
		return
	}

	vnic, err := oracle.GetVnic(pv, pv.Metadata.VnicOcid)
	if err != nil {
		return
	}

	if !vnic.SkipSourceDestCheck {
		err = vnic.SetSkipSourceDestCheck(pv, true)
		if err != nil {
			return
		}
	}

	subnet, err := oracle.GetSubnet(pv, vnic.SubnetId)
	if err != nil {
		return
	}

	tables, err := oracle.GetRouteTables(pv, subnet.VcnId)
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

	curOracleState = blckNet.String()

	return
}
