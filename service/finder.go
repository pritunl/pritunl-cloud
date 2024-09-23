package service

import (
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/shape"
	"regexp"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
)

const (
	DomainKind     = "domain"
	VpcKind        = "vpc"
	DatacenterKind = "datacenter"
	NodeKind       = "node"
	PoolKind       = "pool"
	ZoneKind       = "zone"
	ShapeKind      = "shape"
	ImageKind      = "image"
	InstanceKind   = "instance"
	PlanKind       = "plan"
)

type Resources struct {
	Organization primitive.ObjectID
	Datacenter   *datacenter.Datacenter
	Zone         *zone.Zone
	Vpc          *vpc.Vpc
	Subnet       *vpc.Subnet
	Shape        *shape.Shape
	Node         *node.Node
	Pool         *pool.Pool
	Image        *image.Image
	Instance     *instance.Instance
	Plan         *plan.Plan
	Domain       *domain.Domain
}

var tokenRe = regexp.MustCompile(`{{\.([a-zA-Z0-9-]*)\.([a-zA-Z0-9-]*)}}`)

func (r *Resources) Find(db *database.Database, token string) (err error) {
	matches := tokenRe.FindStringSubmatch(token)
	if len(matches) < 3 {
		err = &errortypes.ParseError{
			errors.Newf("service: Invalid token '%s'", token),
		}
		return
	}

	kind := matches[1]
	resource := matches[2]

	switch kind {
	case DomainKind:
		r.Domain, err = domain.GetOrgName(db, r.Organization, resource)
		if err != nil {
			return
		}
		break
	case VpcKind:
		r.Vpc, err = vpc.GetOrgName(db, r.Organization, resource)
		if err != nil {
			return
		}
		break
	case DatacenterKind:
		r.Datacenter, err = datacenter.GetName(db, resource)
		if err != nil {
			return
		}
		break
	case NodeKind:
		r.Node, err = node.GetName(db, resource)
		if err != nil {
			return
		}
		break
	case PoolKind:
		r.Pool, err = pool.GetName(db, resource)
		if err != nil {
			return
		}
		break
	case ZoneKind:
		r.Zone, err = zone.GetName(db, resource)
		if err != nil {
			return
		}
		break
	case ShapeKind:
		r.Shape, err = shape.GetName(db, resource)
		if err != nil {
			return
		}
		break
	case ImageKind:
		r.Image, err = image.GetOrgPublicName(db, r.Organization, resource)
		if err != nil {
			return
		}
		break
	case InstanceKind:
		r.Instance, err = instance.GetOrgName(db, r.Organization, resource)
		if err != nil {
			return
		}
		break
	case PlanKind:
		r.Plan, err = plan.GetOrgName(db, r.Organization, resource)
		if err != nil {
			return
		}
		break
	default:
		err = &errortypes.ParseError{
			errors.Newf("service: Unknown kind '%s'", kind),
		}
		return
	}

	return
}
