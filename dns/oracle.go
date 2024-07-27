package dns

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/oracle/oci-go-sdk/v65/dns"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/secret"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Oracle struct {
	token       string
	cacheZoneId map[string]string
	provider    *secret.OracleProvider
}

func (o *Oracle) OracleUser() string {
	return ""
}

func (o *Oracle) OraclePrivateKey() string {
	return ""
}

func (o *Oracle) Connect(db *database.Database,
	secr *secret.Secret) (err error) {

	if secr.Type != secret.OracleCloud {
		err = &errortypes.ApiError{
			errors.Wrap(err, "acme: Secret type not Oracle Cloud"),
		}
		return
	}

	o.cacheZoneId = map[string]string{}

	o.provider, err = secr.GetOracleProvider()
	if err != nil {
		return
	}

	return
}

func (o *Oracle) DnsZoneFind(db *database.Database, domain string) (
	zoneId string, err error) {

	domain = extractDomain(domain)

	zoneId = o.cacheZoneId[domain]
	if zoneId != "" {
		return
	}

	compartmentId, err := o.provider.CompartmentOCID()
	if err != nil {
		return
	}

	req := dns.ListZonesRequest{
		CompartmentId: &compartmentId,
	}

	client, err := o.provider.GetDnsClient()
	if err != nil {
		return
	}

	zones, err := client.ListZones(db, req)
	if err != nil {
		err = &errortypes.ApiError{
			errors.Wrap(err, "acme: Oracle zone list error"),
		}
		return
	}

	for _, zone := range zones.Items {
		if matchDomains(*zone.Name, domain) {
			zoneId = *zone.Id
			break
		}
	}

	if zoneId == "" {
		err = &errortypes.ApiError{
			errors.Wrap(err, "acme: Oracle zone not found"),
		}
		return
	}

	o.cacheZoneId[domain] = zoneId

	return
}

func (o *Oracle) DnsTxtGet(db *database.Database,
	domain string) (vals []string, err error) {

	zoneName := extractDomain(domain)
	domain = cleanDomain(domain)

	req := dns.GetZoneRecordsRequest{
		ZoneNameOrId: utils.PointerString(zoneName),
		Domain:       utils.PointerString(domain),
	}

	client, err := o.provider.GetDnsClient()
	if err != nil {
		return
	}

	resp, err := client.GetZoneRecords(db, req)
	if err != nil {
		err = &errortypes.ApiError{
			errors.Wrap(err, "acme: Oracle zone record get error"),
		}
		return
	}

	for _, record := range resp.Items {
		if record.Rtype != nil && *record.Rtype == "TXT" &&
			record.Rdata != nil {

			vals = append(vals, *record.Rdata)
		}
	}

	return
}

func (o *Oracle) DnsTxtUpsert(db *database.Database,
	domain, val string) (err error) {

	zoneName := extractDomain(domain)
	domain = cleanDomain(domain)

	req := dns.PatchZoneRecordsRequest{
		ZoneNameOrId: utils.PointerString(zoneName),
		PatchZoneRecordsDetails: dns.PatchZoneRecordsDetails{
			Items: []dns.RecordOperation{
				{
					Domain: &domain,
					Rtype:  utils.PointerString("TXT"),
					Ttl: utils.PointerInt(
						settings.Acme.DnsOracleCloudTtl),
					Rdata:     utils.PointerString(val),
					Operation: dns.RecordOperationOperationAdd,
				},
			},
		},
	}

	client, err := o.provider.GetDnsClient()
	if err != nil {
		return
	}

	_, err = client.PatchZoneRecords(db, req)
	if err != nil {
		err = &errortypes.ApiError{
			errors.Wrap(err, "acme: Oracle zone patch error"),
		}
		return
	}

	return
}

func (o *Oracle) DnsTxtDelete(db *database.Database,
	domain, val string) (err error) {

	zoneName := extractDomain(domain)
	domain = cleanDomain(domain)

	req := dns.PatchZoneRecordsRequest{
		ZoneNameOrId: utils.PointerString(zoneName),
		PatchZoneRecordsDetails: dns.PatchZoneRecordsDetails{
			Items: []dns.RecordOperation{
				{
					Domain:    &domain,
					Rtype:     utils.PointerString("TXT"),
					Rdata:     utils.PointerString(val),
					Operation: dns.RecordOperationOperationRemove,
				},
			},
		},
	}

	client, err := o.provider.GetDnsClient()
	if err != nil {
		return
	}

	_, err = client.PatchZoneRecords(db, req)
	if err != nil {
		err = &errortypes.ApiError{
			errors.Wrap(err, "acme: Oracle zone patch error"),
		}
		return
	}

	return
}

func (o *Oracle) DnsAGet(db *database.Database,
	domain string) (vals []string, err error) {

	zoneName := extractDomain(domain)
	domain = cleanDomain(domain)

	req := dns.GetZoneRecordsRequest{
		ZoneNameOrId: utils.PointerString(zoneName),
		Domain:       utils.PointerString(domain),
	}

	client, err := o.provider.GetDnsClient()
	if err != nil {
		return
	}

	resp, err := client.GetZoneRecords(db, req)
	if err != nil {
		err = &errortypes.ApiError{
			errors.Wrap(err, "acme: Oracle zone record get error"),
		}
		return
	}

	for _, record := range resp.Items {
		if record.Rtype != nil && *record.Rtype == "A" &&
			record.Rdata != nil {

			vals = append(vals, *record.Rdata)
		}
	}

	return
}

func (o *Oracle) DnsAUpsert(db *database.Database,
	domain, val string) (err error) {

	zoneName := extractDomain(domain)
	domain = cleanDomain(domain)

	req := dns.PatchZoneRecordsRequest{
		ZoneNameOrId: utils.PointerString(zoneName),
		PatchZoneRecordsDetails: dns.PatchZoneRecordsDetails{
			Items: []dns.RecordOperation{
				{
					Domain: &domain,
					Rtype:  utils.PointerString("A"),
					Ttl: utils.PointerInt(
						settings.Acme.DnsOracleCloudTtl),
					Rdata:     utils.PointerString(val),
					Operation: dns.RecordOperationOperationAdd,
				},
			},
		},
	}

	client, err := o.provider.GetDnsClient()
	if err != nil {
		return
	}

	_, err = client.PatchZoneRecords(db, req)
	if err != nil {
		err = &errortypes.ApiError{
			errors.Wrap(err, "acme: Oracle zone patch error"),
		}
		return
	}

	return
}

func (o *Oracle) DnsADelete(db *database.Database,
	domain, val string) (err error) {

	zoneName := extractDomain(domain)
	domain = cleanDomain(domain)

	req := dns.PatchZoneRecordsRequest{
		ZoneNameOrId: utils.PointerString(zoneName),
		PatchZoneRecordsDetails: dns.PatchZoneRecordsDetails{
			Items: []dns.RecordOperation{
				{
					Domain:    &domain,
					Rtype:     utils.PointerString("A"),
					Rdata:     utils.PointerString(val),
					Operation: dns.RecordOperationOperationRemove,
				},
			},
		},
	}

	client, err := o.provider.GetDnsClient()
	if err != nil {
		return
	}

	_, err = client.PatchZoneRecords(db, req)
	if err != nil {
		err = &errortypes.ApiError{
			errors.Wrap(err, "acme: Oracle zone patch error"),
		}
		return
	}

	return
}

func (o *Oracle) DnsAAAAGet(db *database.Database,
	domain string) (vals []string, err error) {

	zoneName := extractDomain(domain)
	domain = cleanDomain(domain)

	req := dns.GetZoneRecordsRequest{
		ZoneNameOrId: utils.PointerString(zoneName),
		Domain:       utils.PointerString(domain),
	}

	client, err := o.provider.GetDnsClient()
	if err != nil {
		return
	}

	resp, err := client.GetZoneRecords(db, req)
	if err != nil {
		err = &errortypes.ApiError{
			errors.Wrap(err, "acme: Oracle zone record get error"),
		}
		return
	}

	for _, record := range resp.Items {
		if record.Rtype != nil && *record.Rtype == "AAAA" &&
			record.Rdata != nil {

			vals = append(vals, *record.Rdata)
		}
	}

	return
}

func (o *Oracle) DnsAAAAUpsert(db *database.Database,
	domain, val string) (err error) {

	zoneName := extractDomain(domain)
	domain = cleanDomain(domain)

	req := dns.PatchZoneRecordsRequest{
		ZoneNameOrId: utils.PointerString(zoneName),
		PatchZoneRecordsDetails: dns.PatchZoneRecordsDetails{
			Items: []dns.RecordOperation{
				{
					Domain: &domain,
					Rtype:  utils.PointerString("AAAA"),
					Ttl: utils.PointerInt(
						settings.Acme.DnsOracleCloudTtl),
					Rdata:     utils.PointerString(val),
					Operation: dns.RecordOperationOperationAdd,
				},
			},
		},
	}

	client, err := o.provider.GetDnsClient()
	if err != nil {
		return
	}

	_, err = client.PatchZoneRecords(db, req)
	if err != nil {
		err = &errortypes.ApiError{
			errors.Wrap(err, "acme: Oracle zone patch error"),
		}
		return
	}

	return
}

func (o *Oracle) DnsAAAADelete(db *database.Database,
	domain, val string) (err error) {

	zoneName := extractDomain(domain)
	domain = cleanDomain(domain)

	req := dns.PatchZoneRecordsRequest{
		ZoneNameOrId: utils.PointerString(zoneName),
		PatchZoneRecordsDetails: dns.PatchZoneRecordsDetails{
			Items: []dns.RecordOperation{
				{
					Domain:    &domain,
					Rtype:     utils.PointerString("AAAA"),
					Rdata:     utils.PointerString(val),
					Operation: dns.RecordOperationOperationRemove,
				},
			},
		},
	}

	client, err := o.provider.GetDnsClient()
	if err != nil {
		return
	}

	_, err = client.PatchZoneRecords(db, req)
	if err != nil {
		err = &errortypes.ApiError{
			errors.Wrap(err, "acme: Oracle zone patch error"),
		}
		return
	}

	return
}
