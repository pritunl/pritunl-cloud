package policy

import (
	"fmt"
	"net"
	"net/http"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/subscription"
	"github.com/pritunl/pritunl-cloud/user"
	"github.com/pritunl/pritunl-cloud/useragent"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

type Rule struct {
	Type    string   `bson:"type" json:"type"`
	Disable bool     `bson:"disable" json:"disable"`
	Values  []string `bson:"values" json:"values"`
}

type Policy struct {
	Id                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name                 string             `bson:"name" json:"name"`
	Comment              string             `bson:"comment" json:"comment"`
	Disabled             bool               `bson:"disabled" json:"disabled"`
	Roles                []string           `bson:"roles" json:"roles"`
	Rules                map[string]*Rule   `bson:"rules" json:"rules"`
	AdminSecondary       primitive.ObjectID `bson:"admin_secondary,omitempty" json:"admin_secondary"`
	UserSecondary        primitive.ObjectID `bson:"user_secondary,omitempty" json:"user_secondary"`
	AdminDeviceSecondary bool               `bson:"admin_device_secondary" json:"admin_device_secondary"`
	UserDeviceSecondary  bool               `bson:"user_device_secondary" json:"user_device_secondary"`
}

func (p *Policy) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	p.Name = utils.FilterName(p.Name)

	if p.Roles == nil {
		p.Roles = []string{}
	}
	if p.Rules == nil {
		p.Rules = map[string]*Rule{}
	}

	for _, rule := range p.Rules {
		switch rule.Type {
		case OperatingSystem:
			break
		case Browser:
			break
		case Location:
			if !subscription.Sub.Active {
				errData = &errortypes.ErrorData{
					Error: "location_subscription_required",
					Message: "Location policy requires subscription " +
						"for GeoIP service.",
				}
				return
			}
			break
		case WhitelistNetworks:
			break
		case BlacklistNetworks:
			break
		default:
			errData = &errortypes.ErrorData{
				Error:   "invalid_rule_type",
				Message: "Rule type is invalid",
			}
			return
		}
	}

	if !p.AdminSecondary.IsZero() &&
		settings.Auth.GetSecondaryProvider(p.AdminSecondary) == nil {

		p.AdminSecondary = primitive.NilObjectID
	}
	if !p.UserSecondary.IsZero() &&
		settings.Auth.GetSecondaryProvider(p.UserSecondary) == nil {

		p.UserSecondary = primitive.NilObjectID
	}

	hasWebAuthn := false
	nodes, err := node.GetAll(db)
	if err != nil {
		return
	}

	for _, nde := range nodes {
		if nde.WebauthnDomain != "" {
			hasWebAuthn = true
			break
		}
	}

	if (p.AdminDeviceSecondary || p.UserDeviceSecondary) && !hasWebAuthn {
		errData = &errortypes.ErrorData{
			Error: "webauthn_domain_unavailable",
			Message: "At least one node must have a WebAuthn domain " +
				"configured to use WebAuthn device authentication",
		}
		return
	}

	return
}

func (p *Policy) ValidateUser(db *database.Database, usr *user.User,
	r *http.Request) (errData *errortypes.ErrorData, err error) {

	if p.Disabled {
		return
	}

	agnt, err := useragent.Parse(db, r)
	if err != nil {
		return
	}

	for _, rule := range p.Rules {
		switch rule.Type {
		case OperatingSystem:
			match := false
			for _, value := range rule.Values {
				if value == agnt.OperatingSystem {
					match = true
					break
				}
			}

			if !match {
				if rule.Disable {
					errData = &errortypes.ErrorData{
						Error:   "unauthorized",
						Message: "Not authorized",
					}

					usr.Disabled = true
					err = usr.CommitFields(db, set.NewSet("disabled"))
					if err != nil {
						return
					}
				} else {
					errData = &errortypes.ErrorData{
						Error:   "operating_system_policy",
						Message: "Operating system not permitted",
					}
				}
				return
			}
			break
		case Browser:
			match := false
			for _, value := range rule.Values {
				if value == agnt.Browser {
					match = true
					break
				}
			}

			if !match {
				if rule.Disable {
					errData = &errortypes.ErrorData{
						Error:   "unauthorized",
						Message: "Not authorized",
					}

					usr.Disabled = true
					err = usr.CommitFields(db, set.NewSet("disabled"))
					if err != nil {
						return
					}
				} else {
					errData = &errortypes.ErrorData{
						Error:   "browser_policy",
						Message: "Browser not permitted",
					}
				}
				return
			}
			break
		case Location:
			match := false
			regionKey := fmt.Sprintf("%s_%s",
				agnt.CountryCode, agnt.RegionCode)

			for _, value := range rule.Values {
				if value == agnt.CountryCode || value == regionKey {
					match = true
					break
				}
			}

			if !match {
				if rule.Disable {
					errData = &errortypes.ErrorData{
						Error:   "unauthorized",
						Message: "Not authorized",
					}

					usr.Disabled = true
					err = usr.CommitFields(db, set.NewSet("disabled"))
					if err != nil {
						return
					}
				} else {
					errData = &errortypes.ErrorData{
						Error:   "location_policy",
						Message: "Location not permitted",
					}
				}
				return
			}
			break
		case WhitelistNetworks:
			match := false
			clientIp := net.ParseIP(agnt.Ip)

			for _, value := range rule.Values {
				_, network, e := net.ParseCIDR(value)
				if e != nil {
					err = &errortypes.ParseError{
						errors.Wrap(e, "policy: Failed to parse network"),
					}

					logrus.WithFields(logrus.Fields{
						"network": value,
						"error":   err,
					}).Error("policy: Invalid whitelist network")
					err = nil
					continue
				}

				if network.Contains(clientIp) {
					match = true
					break
				}
			}

			if !match {
				if rule.Disable {
					errData = &errortypes.ErrorData{
						Error:   "unauthorized",
						Message: "Not authorized",
					}

					usr.Disabled = true
					err = usr.CommitFields(db, set.NewSet("disabled"))
					if err != nil {
						return
					}
				} else {
					errData = &errortypes.ErrorData{
						Error:   "whitelist_networks_policy",
						Message: "Network not permitted",
					}
				}
				return
			}
			break
		case BlacklistNetworks:
			match := false
			clientIp := net.ParseIP(agnt.Ip)

			for _, value := range rule.Values {
				_, network, e := net.ParseCIDR(value)
				if e != nil {
					err = &errortypes.ParseError{
						errors.Wrap(e, "policy: Failed to parse network"),
					}

					logrus.WithFields(logrus.Fields{
						"network": value,
						"error":   err,
					}).Error("policy: Invalid blacklist network")
					err = nil
					continue
				}

				if network.Contains(clientIp) {
					match = true
					break
				}
			}

			if match {
				if rule.Disable {
					errData = &errortypes.ErrorData{
						Error:   "unauthorized",
						Message: "Not authorized",
					}

					usr.Disabled = true
					err = usr.CommitFields(db, set.NewSet("disabled"))
					if err != nil {
						return
					}
				} else {
					errData = &errortypes.ErrorData{
						Error:   "blacklist_networks_policy",
						Message: "Network not permitted",
					}
				}
				return
			}
			break
		}
	}

	return
}

func (p *Policy) Commit(db *database.Database) (err error) {
	coll := db.Policies()

	err = coll.Commit(p.Id, p)
	if err != nil {
		return
	}

	return
}

func (p *Policy) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Policies()

	err = coll.CommitFields(p.Id, p, fields)
	if err != nil {
		return
	}

	return
}

func (p *Policy) Insert(db *database.Database) (err error) {
	coll := db.Policies()

	if !p.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("policy: Policy already exists"),
		}
		return
	}

	_, err = coll.InsertOne(db, p)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
