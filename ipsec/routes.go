package ipsec

import (
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/link"
	"github.com/pritunl/pritunl-cloud/vpc"
	"strings"
)

func addRoutes(db *database.Database, vc *vpc.Vpc, states []*link.State,
	addr, addr6 string) (err error) {

	routes := []*vpc.Route{}

	for _, state := range states {
		for _, lnk := range state.Links {
			for _, dst := range lnk.RightSubnets {
				if strings.Contains(dst, ":") {
					routes = append(routes, &vpc.Route{
						Destination: dst,
						Target:      addr6,
						Link:        true,
					})
				} else {
					routes = append(routes, &vpc.Route{
						Destination: dst,
						Target:      addr,
						Link:        true,
					})
				}
			}
		}
	}

	err = vc.AddLinkRoutes(db, routes)
	if err != nil {
		return
	}

	return
}
