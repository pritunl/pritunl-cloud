package spec

import (
	"regexp"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/domain"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/image"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/plan"
	"github.com/pritunl/pritunl-cloud/pool"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/shape"
	"github.com/pritunl/pritunl-cloud/vpc"
	"github.com/pritunl/pritunl-cloud/zone"
)

const (
	DomainKind      = "domain"
	VpcKind         = "vpc"
	SubnetKind      = "subnet"
	DatacenterKind  = "datacenter"
	NodeKind        = "node"
	PoolKind        = "pool"
	ZoneKind        = "zone"
	ShapeKind       = "shape"
	ImageKind       = "image"
	InstanceKind    = "instance"
	PlanKind        = "plan"
	CertificateKind = "certificate"
	SecretKind      = "secret"
	PodKind         = "pod"
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
	Certificate  *certificate.Certificate
	Secret       *secret.Secret
	Pod          *PodBase
}

var tokenRe = regexp.MustCompile(`{{\/([a-zA-Z0-9-]*)\/([a-zA-Z0-9-]*)}}`)

func (r *Resources) Find(db *database.Database, token string) (
	kind string, err error) {

	matches := tokenRe.FindStringSubmatch(token)
	if len(matches) < 3 {
		err = &errortypes.ParseError{
			errors.Newf("spec: Invalid token '%s'", token),
		}
		return
	}

	kind = matches[1]
	resource := matches[2]

	switch kind {
	case DomainKind:
		r.Domain, err = domain.GetOne(db, &bson.M{
			"name":         resource,
			"organization": r.Organization,
		})
		if err != nil {
			return
		}
		break
	case VpcKind:
		r.Vpc, err = vpc.GetOne(db, &bson.M{
			"name":         resource,
			"organization": r.Organization,
		})
		if err != nil {
			return
		}
		break
	case SubnetKind:
		if r.Vpc != nil {
			subnet := r.Vpc.GetSubnetName(resource)
			r.Subnet = subnet
		}
		break
	case DatacenterKind:
		r.Datacenter, err = datacenter.GetOne(db, &bson.M{
			"name": resource,
		})
		if err != nil {
			return
		}
		break
	case NodeKind:
		r.Node, err = node.GetOne(db, &bson.M{
			"name": resource,
		})
		if err != nil {
			return
		}
		break
	case PoolKind:
		r.Pool, err = pool.GetOne(db, &bson.M{
			"name": resource,
		})
		if err != nil {
			return
		}
		break
	case ZoneKind:
		r.Zone, err = zone.GetOne(db, &bson.M{
			"name": resource,
		})
		if err != nil {
			return
		}
		break
	case ShapeKind:
		query := bson.M{
			"name": resource,
		}
		if r.Datacenter != nil {
			query["datacenter"] = r.Datacenter.Id
		} else if r.Zone != nil {
			query["datacenter"] = r.Zone.Datacenter
		}
		r.Shape, err = shape.GetOne(db, &query)
		if err != nil {
			return
		}
		break
	case ImageKind:
		r.Image, err = image.GetOne(db, &bson.M{
			"name": resource,
			"$or": []*bson.M{
				&bson.M{
					"organization": r.Organization,
				},
				&bson.M{
					"organization": &bson.M{
						"$exists": false,
					},
				},
			},
		})
		if err != nil {
			return
		}
		break
	case InstanceKind:
		r.Instance, err = instance.GetOne(db, &bson.M{
			"name":         resource,
			"organization": r.Organization,
		})
		if err != nil {
			return
		}
		break
	case PlanKind:
		r.Plan, err = plan.GetOne(db, &bson.M{
			"name":         resource,
			"organization": r.Organization,
		})
		if err != nil {
			return
		}
		break
	case CertificateKind:
		r.Certificate, err = certificate.GetOne(db, &bson.M{
			"name":         resource,
			"organization": r.Organization,
		})
		if err != nil {
			return
		}
		break
	case SecretKind:
		r.Secret, err = secret.GetOne(db, &bson.M{
			"name":         resource,
			"organization": r.Organization,
		})
		if err != nil {
			return
		}
		break
	case PodKind:
		r.Pod, err = GetPodBase(db, &bson.M{
			"name":         resource,
			"organization": r.Organization,
		})
		if err != nil {
			return
		}
		break
	default:
		err = &errortypes.ParseError{
			errors.Newf("spec: Unknown kind '%s'", kind),
		}
		return
	}

	return
}

type PodBase struct {
	Id   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
}

func GetPodBase(db *database.Database, query *bson.M) (
	pd *PodBase, err error) {

	coll := db.Pods()
	pd = &PodBase{}

	err = coll.FindOne(db, query).Decode(pd)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
