package finder

import (
	"regexp"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/disk"
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

type Resources struct {
	Organization bson.ObjectID
	Datacenter   *datacenter.Datacenter
	Zone         *zone.Zone
	Vpc          *vpc.Vpc
	Subnet       *vpc.Subnet
	Shape        *shape.Shape
	Node         *node.Node
	Pool         *pool.Pool
	Image        *image.Image
	Disks        []*disk.Disk
	Instance     *instance.Instance
	Plan         *plan.Plan
	Domain       *domain.Domain
	Certificate  *certificate.Certificate
	Secret       *secret.Secret
	Deployment   *deployment.Deployment
	Pod          *PodBase
	Unit         *UnitBase
	Selector     string
}

var tokenRe = regexp.MustCompile(
	`\+\/([a-zA-Z0-9-]*)\/([a-zA-Z0-9-_.]*)(?:(?:\/|\:)([a-zA-Z0-9-_.]*)(?:\/([a-zA-Z0-9-_.]*))?)?`)

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
	tag := ""
	r.Selector = ""

	if len(matches) > 3 {
		if strings.Contains(token, ":") {
			tag = matches[3]
			if len(matches) > 4 {
				r.Selector = matches[4]
			}
		} else {
			r.Selector = matches[3]
		}
	}

	switch kind {
	case DomainKind:
		r.Domain, err = domain.GetOne(db, &bson.M{
			"name":         resource,
			"organization": r.Organization,
		})
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}
		break
	case VpcKind:
		query := bson.M{
			"name":         resource,
			"organization": r.Organization,
		}
		if r.Datacenter != nil {
			query["datacenter"] = r.Datacenter.Id
		} else if r.Zone != nil {
			query["datacenter"] = r.Zone.Datacenter
		}
		r.Vpc, err = vpc.GetOne(db, &query)
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
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
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}
		break
	case NodeKind:
		r.Node, err = node.GetOne(db, &bson.M{
			"name": resource,
		})
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}
		break
	case PoolKind:
		r.Pool, err = pool.GetOne(db, &bson.M{
			"name": resource,
		})
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}
		break
	case ZoneKind:
		r.Zone, err = zone.GetOne(db, &bson.M{
			"name": resource,
		})
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}

		r.Datacenter, err = datacenter.Get(db, r.Zone.Datacenter)
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
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
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}
		break
	case ImageKind:
		if image.Releases.Contains(resource) {
			imgs, e := image.GetAll(db, &bson.M{
				"release": resource,
				"organization": &bson.M{
					"$in": []bson.ObjectID{r.Organization, image.Global},
				},
			})
			if e != nil {
				err = e
				if _, ok := err.(*database.NotFoundError); ok {
					err = nil
				} else {
					return
				}
			}

			var latestImg *image.Image
			for _, img := range imgs {
				if latestImg == nil {
					latestImg = img
				} else if img.Build > latestImg.Build {
					latestImg = img
				}

				if img.Build == tag {
					r.Image = img
					break
				}
			}

			if latestImg != nil && (tag == "" || tag == "latest") {
				r.Image = latestImg
			}
		}

		if r.Image == nil {
			r.Image, err = image.GetOne(db, &bson.M{
				"name": resource,
				"organization": &bson.M{
					"$in": []bson.ObjectID{r.Organization, image.Global},
				},
			})
			if err != nil {
				if _, ok := err.(*database.NotFoundError); ok {
					err = nil
				} else {
					return
				}
			}
		}

		break
	case BuildKind:
		r.Unit, err = GetUnitBase(db, r.Organization, resource)
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}

		if r.Unit != nil {
			if tag == "" || tag == "latest" {
				deplys, e := deployment.GetAllSorted(db, &bson.M{
					"unit": r.Unit.Id,
				})
				if e != nil {
					err = e
					if _, ok := err.(*database.NotFoundError); ok {
						err = nil
					} else {
						return
					}
				}

				for _, deply := range deplys {
					r.Deployment = deply
					break
				}
			} else {
				deplys, e := deployment.GetAllSorted(db, &bson.M{
					"unit": r.Unit.Id,
					"tags": tag,
				})
				if e != nil {
					err = e
					if _, ok := err.(*database.NotFoundError); ok {
						err = nil
					} else {
						return
					}
				}

				for _, deply := range deplys {
					r.Deployment = deply
					break
				}
			}
		}
		break
	case DiskKind:
		r.Disks, err = disk.GetAll(db, &bson.M{
			"name":         resource,
			"organization": r.Organization,
		})
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}
		break
	case InstanceKind:
		r.Instance, err = instance.GetOne(db, &bson.M{
			"name":         resource,
			"organization": r.Organization,
		})
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}
		break
	case PlanKind:
		r.Plan, err = plan.GetOne(db, &bson.M{
			"name":         resource,
			"organization": r.Organization,
		})
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}
		break
	case CertificateKind:
		r.Certificate, err = certificate.GetOne(db, &bson.M{
			"name":         resource,
			"organization": r.Organization,
		})
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}
		break
	case SecretKind:
		r.Secret, err = secret.GetOne(db, &bson.M{
			"name":         resource,
			"organization": r.Organization,
		})
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}
		break
	case PodKind:
		r.Pod, err = GetPodBase(db, &bson.M{
			"name":         resource,
			"organization": r.Organization,
		})
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
		}
		break
	case UnitKind:
		r.Unit, err = GetUnitBase(db, r.Organization, resource)
		if err != nil {
			if _, ok := err.(*database.NotFoundError); ok {
				err = nil
			} else {
				return
			}
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
	Id           bson.ObjectID `bson:"_id,omitempty"`
	Organization bson.ObjectID `bson:"organization"`
	Name         string        `bson:"name"`
}

type UnitBase struct {
	Id           bson.ObjectID `bson:"_id,omitempty"`
	Pod          bson.ObjectID `bson:"pod"`
	Organization bson.ObjectID `bson:"organization"`
	Name         string        `bson:"name"`
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

func GetUnitBase(db *database.Database, orgId bson.ObjectID,
	name string) (unt *UnitBase, err error) {

	coll := db.Units()
	unt = &UnitBase{}

	err = coll.FindOne(db, &bson.M{
		"name":         name,
		"organization": orgId,
	}).Decode(unt)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
