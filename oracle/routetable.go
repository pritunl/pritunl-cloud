package oracle

import (
	"context"

	"github.com/dropbox/godropbox/errors"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type RouteTable struct {
	Id         string
	VcnId      string
	Routes     map[string]string
	routeRules []core.RouteRule
}

func (r *RouteTable) RouteExists(dest string, nextHopId string) bool {
	if r.Routes[dest] == nextHopId {
		return true
	}
	return false
}

func (r *RouteTable) RouteUpsert(dest string, nextHopId string) bool {
	for i, routeRule := range r.routeRules {
		if routeRule.Destination != nil &&
			*routeRule.Destination == dest {

			if routeRule.NetworkEntityId != nil &&
				*routeRule.NetworkEntityId != nextHopId {

				routeRule.NetworkEntityId = &nextHopId
				r.routeRules[i] = routeRule
				return true
			} else {
				return false
			}
		}
	}

	routeRule := core.RouteRule{
		Destination:     &dest,
		NetworkEntityId: &nextHopId,
	}
	r.routeRules = append(r.routeRules, routeRule)
	return true
}

func (r *RouteTable) CommitRouteRules(pv *Provider) (err error) {
	client, err := pv.GetNetworkClient()
	if err != nil {
		return
	}

	req := core.UpdateRouteTableRequest{
		RtId: &r.Id,
		UpdateRouteTableDetails: core.UpdateRouteTableDetails{
			RouteRules: r.routeRules,
		},
	}

	_, err = client.UpdateRouteTable(context.Background(), req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "oracle: Failed to update route table"),
		}
		return
	}

	return
}

func GetRouteTables(pv *Provider, vcnId string) (
	tables []*RouteTable, err error) {

	limit := 100
	compartmentId, err := pv.CompartmentOCID()
	if err != nil {
		return
	}

	client, err := pv.GetNetworkClient()
	if err != nil {
		return
	}

	vnicReq := core.ListRouteTablesRequest{
		CompartmentId: &compartmentId,
		VcnId:         &vcnId,
		Limit:         &limit,
	}

	orcTables, err := client.ListRouteTables(context.Background(), vnicReq)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "oracle: Failed to list route tables"),
		}
		return
	}

	tables = []*RouteTable{}
	if orcTables.Items != nil {
		for _, orcTable := range orcTables.Items {
			table := &RouteTable{}

			if orcTable.Id != nil {
				table.Id = *orcTable.Id
			}
			if orcTable.VcnId != nil {
				table.VcnId = *orcTable.VcnId
			}
			if orcTable.RouteRules != nil {
				table.routeRules = orcTable.RouteRules
			} else {
				table.routeRules = []core.RouteRule{}
			}

			routes := map[string]string{}
			for _, rule := range table.routeRules {
				if rule.Destination == nil || rule.NetworkEntityId == nil {
					continue
				}

				routes[*rule.Destination] = *rule.NetworkEntityId
			}
			table.Routes = routes

			tables = append(tables, table)
		}
	}

	return
}
