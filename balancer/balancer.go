package balancer

import (
	"fmt"
	"net"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/utils"
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
	WebSockets  int       `bson:"websockets" json:"websockets"`
	Online      []string  `bson:"online" json:"online"`
	UnknownHigh []string  `bson:"unknown_high" json:"unknown_high"`
	UnknownMid  []string  `bson:"unknown_mid" json:"unknown_mid"`
	UnknownLow  []string  `bson:"unknown_low" json:"unknown_low"`
	Offline     []string  `bson:"offline" json:"offline"`
}

type Balancer struct {
	Id              bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name            string            `bson:"name" json:"name"`
	Comment         string            `bson:"comment" json:"comment"`
	Type            string            `bson:"type" json:"type"`
	State           bool              `bson:"state" json:"state"`
	Organization    bson.ObjectID     `bson:"organization" json:"organization"`
	Datacenter      bson.ObjectID     `bson:"datacenter" json:"datacenter"`
	Certificates    []bson.ObjectID   `bson:"certificates" json:"certificates"`
	ClientAuthority bson.ObjectID     `bson:"client_authority" json:"client_authority"`
	WebSockets      bool              `bson:"websockets" json:"websockets"`
	Domains         []*Domain         `bson:"domains" json:"domains"`
	Backends        []*Backend        `bson:"backends" json:"backends"`
	States          map[string]*State `bson:"states" json:"states"`
	CheckPath       string            `bson:"check_path" json:"check_path"`
}

func (b *Balancer) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	b.Name = utils.FilterName(b.Name)

	if b.Type == "" {
		b.Type = Http
	}

	if b.Domains == nil {
		b.Domains = []*Domain{}
	}

	domains := []*Domain{}
	for _, domain := range b.Domains {
		domain.Domain = utils.FilterDomain(domain.Domain)
		domain.Host = utils.FilterDomain(domain.Host)

		if domain.Domain == "" {
			continue
		}

		domains = append(domains, domain)
	}
	b.Domains = domains

	if b.Backends == nil {
		b.Backends = []*Backend{}
	}

	if b.Certificates == nil {
		b.Certificates = []bson.ObjectID{}
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

		if backend.Port < 1 || backend.Port > 65535 {
			errData = &errortypes.ErrorData{
				Error:   "balancer_port_invalid",
				Message: "Invalid balancer backend port",
			}
			return
		}

		ip := net.ParseIP(backend.Hostname)
		if ip == nil {
			errData = &errortypes.ErrorData{
				Error: "balancer_hostname_invalid",
				Message: fmt.Sprintf("Balancer hostname '%s' must "+
					"match existing instance address", backend.Hostname),
			}
			return
		}
		backend.Hostname = ip.String()

		exists, e := instance.ExistsIp(db, backend.Hostname)
		if e != nil {
			err = e
			return
		}

		if !exists {
			errData = &errortypes.ErrorData{
				Error: "balancer_hostname_not_found",
				Message: fmt.Sprintf("Balancer hostname '%s' must "+
					"match existing instance address", backend.Hostname),
			}
			return
		}
	}

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

	if b.State {
		if len(b.Domains) == 0 {
			errData = &errortypes.ErrorData{
				Error:   "domain_required",
				Message: "Missing required domain",
			}
			return
		}

		if b.CheckPath == "" {
			errData = &errortypes.ErrorData{
				Error:   "check_path_required",
				Message: "Missing required health check path",
			}
			return
		}

		if len(b.Backends) == 0 {
			errData = &errortypes.ErrorData{
				Error:   "backend_required",
				Message: "Missing required backend",
			}
			return
		}

		domains := []string{}
		for _, domain := range b.Domains {
			domains = append(domains, domain.Domain)
		}

		coll := db.Balancers()
		count, e := coll.CountDocuments(db, &bson.M{
			"_id": &bson.M{
				"$ne": b.Id,
			},
			"state": true,
			"domains.domain": &bson.M{
				"$in": domains,
			},
		})
		if e != nil {
			err = database.ParseError(e)
			return
		}

		if count > 0 {
			errData = &errortypes.ErrorData{
				Error: "domain_conflict",
				Message: "External domain conflicts with another " +
					"load balancer in same datacenter",
			}
			return
		}
	}

	return
}

func (b *Balancer) Json() {
	if b.States == nil || len(b.States) == 0 {
		return
	}

	for key, state := range b.States {
		if time.Since(state.Timestamp) > 1*time.Minute {
			delete(b.States, key)
		}
	}

	return
}

func (b *Balancer) Clean(db *database.Database) (err error) {
	if b.States == nil || len(b.States) == 0 {
		return
	}

	changed := false
	for key, state := range b.States {
		if time.Since(state.Timestamp) > 1*time.Minute {
			changed = true
			delete(b.States, key)
		}
	}

	if changed {
		err = b.CommitFields(db, set.NewSet("states"))
		if err != nil {
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

	if b.State && (fields.Contains("state") || fields.Contains("domains")) {
		domains := []string{}
		for _, domain := range b.Domains {
			domains = append(domains, domain.Domain)
		}

		coll := db.Balancers()
		count, e := coll.CountDocuments(db, &bson.M{
			"_id": &bson.M{
				"$ne": b.Id,
			},
			"state": true,
			"domains.domain": &bson.M{
				"$in": domains,
			},
		})
		if e != nil {
			err = database.ParseError(e)
			return
		}

		if count > 0 {
			err = &errortypes.ReadError{
				errors.New("balancer: Datacenter domain conflict"),
			}
			return
		}
	}

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
