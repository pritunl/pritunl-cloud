package dns

import (
	"github.com/cloudflare/cloudflare-go"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"strings"
)

type Cloudflare struct {
	sess        *cloudflare.API
	token       string
	cacheZoneId map[string]string
}

func (c *Cloudflare) Connect(db *database.Database,
	secr *secret.Secret) (err error) {

	if secr.Type != secret.Cloudflare {
		err = &errortypes.ApiError{
			errors.Wrap(err, "acme: Secret type not cloudflare"),
		}
		return
	}

	c.sess, err = cloudflare.NewWithAPIToken(utils.FilterStr(secr.Key, 256))
	if err != nil {
		err = &errortypes.ApiError{
			errors.Wrap(err, "dns: Failed to connect to cloudflare api"),
		}
		return
	}

	c.cacheZoneId = map[string]string{}

	return
}

func (c *Cloudflare) DnsZoneFind(db *database.Database, domain string) (
	zoneId string, err error) {

	domain = extractDomain(domain)

	zoneId = c.cacheZoneId[domain]
	if zoneId != "" {
		return
	}

	zones, err := c.sess.ListZones(db)
	if err != nil {
		err = &errortypes.ApiError{
			errors.Wrap(err, "acme: Failed to list cloudflare zones"),
		}
		return
	}

	for _, zone := range zones {
		if strings.Trim(zone.Name, ".") == strings.Trim(domain, ".") {
			zoneId = zone.ID
			break
		}
	}

	if zoneId == "" {
		err = &errortypes.ApiError{
			errors.Wrap(err, "acme: Cloudflare zone not found"),
		}
		return
	}

	c.cacheZoneId[domain] = zoneId

	return
}

func (c *Cloudflare) DnsTxtGet(db *database.Database,
	domain string) (vals []string, err error) {

	vals = []string{}

	domain = cleanDomain(domain)

	zoneId, err := c.DnsZoneFind(db, domain)
	if err != nil {
		return
	}

	listParams := cloudflare.ListDNSRecordsParams{
		Type: "TXT",
		Name: domain,
	}

	records, _, err := c.sess.ListDNSRecords(
		db,
		cloudflare.ZoneIdentifier(zoneId),
		listParams,
	)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "acme: Failed to get DNS records"),
		}
		return
	}

	for _, record := range records {
		if record.Type == "TXT" &&
			strings.Trim(record.Name, ".") == strings.Trim(domain, ".") {

			vals = append(vals, record.Content)
			break
		}
	}

	return
}

func (c *Cloudflare) DnsTxtUpsert(db *database.Database,
	domain, val string) (err error) {

	domain = cleanDomain(domain)

	zoneId, err := c.DnsZoneFind(db, domain)
	if err != nil {
		return
	}

	listParams := cloudflare.ListDNSRecordsParams{
		Type: "TXT",
		Name: domain,
	}

	records, _, err := c.sess.ListDNSRecords(
		db,
		cloudflare.ZoneIdentifier(zoneId),
		listParams,
	)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "acme: Failed to get DNS records"),
		}
		return
	}

	recordId := ""
	for _, record := range records {
		if record.Type == "TXT" &&
			strings.Trim(record.Name, ".") == strings.Trim(domain, ".") {

			recordId = record.ID
			break
		}
	}

	if recordId == "" {
		createParams := cloudflare.CreateDNSRecordParams{
			Type:    "TXT",
			Name:    domain,
			Content: val,
			TTL:     settings.Acme.DnsCloudflareTtl,
		}

		_, err = c.sess.CreateDNSRecord(
			db,
			cloudflare.ZoneIdentifier(zoneId),
			createParams,
		)
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrap(err, "acme: Failed to create DNS record"),
			}
			return
		}
	} else {
		updateParams := cloudflare.UpdateDNSRecordParams{
			Type:    "TXT",
			Name:    domain,
			Content: val,
			TTL:     settings.Acme.DnsCloudflareTtl,
		}

		_, err = c.sess.UpdateDNSRecord(
			db,
			cloudflare.ResourceIdentifier(recordId),
			updateParams,
		)
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrap(err, "acme: Failed to update DNS record"),
			}
			return
		}
	}

	return
}

func (c *Cloudflare) DnsTxtDelete(db *database.Database,
	domain, val string) (err error) {

	domain = cleanDomain(domain)

	zoneId, err := c.DnsZoneFind(db, domain)
	if err != nil {
		return
	}

	listParams := cloudflare.ListDNSRecordsParams{
		Type: "TXT",
		Name: domain,
	}

	records, _, err := c.sess.ListDNSRecords(
		db,
		cloudflare.ZoneIdentifier(zoneId),
		listParams,
	)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "acme: Failed to get DNS records"),
		}
		return
	}

	recordId := ""
	for _, record := range records {
		if record.Type == "TXT" &&
			strings.Trim(record.Name, ".") == strings.Trim(domain, ".") &&
			strings.ReplaceAll(record.Content, "\"",
				"") == strings.ReplaceAll(val, "\"", "") {

			recordId = record.ID
			break
		}
	}

	if recordId != "" {
		err = c.sess.DeleteDNSRecord(
			db,
			cloudflare.ZoneIdentifier(zoneId),
			recordId,
		)
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrap(err, "acme: Failed to delete DNS record"),
			}
			return
		}
	}

	return
}
