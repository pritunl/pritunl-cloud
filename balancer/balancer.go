package balancer

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
)

type Domain struct {
	Domain string `bson:"domain" json:"domain"`
	Host   string `bson:"host" json:"host"`
}

type Backend struct {
	Protocol string `bson:"protocol" json:"protocol"`
	Hostname string `bson:"hostname" json:"hostname"`
	Port     int    `bson:"port" json:"port"`
}

type State struct {
	Timestamp   time.Time `bson:"timestamp" json:"timestamp"`
	Requests    int       `bson:"requests" json:"requests"`
	Retries     int       `bson:"retries" json:"retries"`
	Online      []string  `bson:"online" json:"online"`
	UnknownHigh []string  `bson:"unknown_high" json:"unknown_high"`
	UnknownMid  []string  `bson:"unknown_mid" json:"unknown_mid"`
	UnknownLow  []string  `bson:"unknown_low" json:"unknown_low"`
	Offline     []string  `bson:"offline" json:"offline"`
}

type Balancer struct {
	Id              primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name            string               `bson:"name" json:"name"`
	Type            string               `bson:"type" json:"type"`
	State           bool                 `bson:"state" json:"state"`
	Organization    primitive.ObjectID   `bson:"organization,omitempty" json:"organization"`
	Datacenter      primitive.ObjectID   `bson:"datacenter,omitempty" json:"datacenter"`
	Certificates    []primitive.ObjectID `bson:"certificates" json:"certificates"`
	ClientAuthority primitive.ObjectID   `bson:"client_authority" json:"client_authority"`
	WebSockets      bool                 `bson:"websockets" json:"websockets"`
	Domains         []*Domain            `bson:"domains" json:"domains"`
	Backends        []*Backend           `bson:"backends" json:"backends"`
	States          map[string]*State    `bson:"states" json:"states"`
}

func (b *Balancer) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if b.Type == "" {
		b.Type = Http
	}

	if b.Domains == nil {
		b.Domains = []*Domain{}
	}

	if b.Backends == nil {
		b.Backends = []*Backend{}
	}

	if b.Certificates == nil {
		b.Certificates = []primitive.ObjectID{}
	}

	if b.States == nil {
		b.States = map[string]*State{}
	}

	for _, backend := range b.Backends {
		if backend.Protocol != "http" && backend.Protocol != "https" {
			errData = &errortypes.ErrorData{
				Error:   "balancer_protocol_invalid",
				Message: "Invalid balancer backend protocol",
			}
			return
		}

		if backend.Hostname == "" {
			errData = &errortypes.ErrorData{
				Error:   "balancer_hostname_invalid",
				Message: "Invalid balancer backend hostname",
			}
			return
		}

		if backend.Port == 0 {
			errData = &errortypes.ErrorData{
				Error:   "balancer_port_invalid",
				Message: "Invalid balancer backend port",
			}
			return
		}
	}

	if b.State {
		if b.Organization.IsZero() {
			errData = &errortypes.ErrorData{
				Error:   "organization_required",
				Message: "Missing required organization",
			}
			return
		}

		if b.Datacenter.IsZero() {
			errData = &errortypes.ErrorData{
				Error:   "datacenter_required",
				Message: "Missing required datacenter",
			}
			return
		}

		if b.Domains == nil || len(b.Domains) == 0 {
			errData = &errortypes.ErrorData{
				Error:   "domain_required",
				Message: "Missing required domain",
			}
			return
		}

		if b.Backends == nil || len(b.Backends) == 0 {
			errData = &errortypes.ErrorData{
				Error:   "backend_required",
				Message: "Missing required backend",
			}
			return
		}
	}

	return
}

func (b *Balancer) CommitState(db *database.Database, state *State) (
	err error) {

	coll := db.Balancers()
	_, err = coll.UpdateOne(db, &bson.M{
		"_id": b.Id,
	}, &bson.M{
		"$set": &bson.M{
			"states." + node.Self.Id.Hex(): state,
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func (b *Balancer) Commit(db *database.Database) (err error) {
	coll := db.Balancers()

	err = coll.Commit(b.Id, b)
	if err != nil {
		return
	}

	return
}

func (b *Balancer) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Balancers()

	err = coll.CommitFields(b.Id, b, fields)
	if err != nil {
		return
	}

	return
}

func (b *Balancer) Insert(db *database.Database) (err error) {
	coll := db.Balancers()

	if !b.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("balancer: Balancer already exists"),
		}
		return
	}

	_, err = coll.InsertOne(db, b)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
