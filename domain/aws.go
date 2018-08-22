package domain

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"strings"
)

type awsProvider struct {
	domain *Domain
}

func (p *awsProvider) Retrieve() (val credentials.Value, err error) {
	val = credentials.Value{
		AccessKeyID:     p.domain.AwsId,
		SecretAccessKey: p.domain.AwsSecret,
	}

	return
}

func (p *awsProvider) IsExpired() bool {
	return false
}

func awsGetSession(domain *Domain) (sess *session.Session, err error) {
	prov := &awsProvider{
		domain: domain,
	}

	cred := credentials.NewCredentials(prov)

	conf := aws.Config{
		Credentials: cred,
	}

	opts := session.Options{
		Config:            conf,
		SharedConfigState: session.SharedConfigEnable,
	}

	sess, err = session.NewSessionWithOptions(opts)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "domain: Failed to create AWS session"),
		}
		return
	}

	return
}

func AwsUpsertDomain(domain *Domain, name, addr, addr6 string) (err error) {
	sess, err := awsGetSession(domain)
	if err != nil {
		return
	}

	servc := route53.New(sess)

	zones, err := servc.ListHostedZonesByName(nil)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "domain: Failed to list Route53 zones"),
		}
		return
	}

	zoneId := ""
	zoneName := ""
	for _, zone := range zones.HostedZones {
		if strings.TrimRight(*zone.Name, ".") != domain.Name {
			continue
		}

		zoneId = *zone.Id
		zoneName = *zone.Name
	}

	if zoneId == "" {
		err = &errortypes.RequestError{
			errors.Wrap(err, "domain: Failed to find Route53 zone"),
		}
		return
	}

	recordName := name + "." + zoneName

	records, err := servc.ListResourceRecordSets(
		&route53.ListResourceRecordSetsInput{
			HostedZoneId: &zoneId,
		},
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "domain: Failed to list Route53 records"),
		}
		return
	}

	curAddrs := []string{}
	curAddrs6 := []string{}
	for _, record := range records.ResourceRecordSets {
		if (*record.Type != "A" && *record.Type != "AAAA") ||
			*record.Name != recordName {

			continue
		}

		for _, resource := range record.ResourceRecords {
			if *record.Type == "A" {
				curAddrs = append(curAddrs, *resource.Value)
			} else if *record.Type == "AAAA" {
				curAddrs6 = append(curAddrs6, *resource.Value)
			}
		}
	}

	changes := []*route53.Change{}

	if len(curAddrs) != 1 || curAddrs[0] != addr {
		if addr == "" && len(curAddrs) > 0 {
			action := "DELETE"
			recordSetType := "A"
			recordSetTtl := int64(60)
			records := []*route53.ResourceRecord{}

			for _, a := range curAddrs {
				records = append(records, &route53.ResourceRecord{
					Value: &a,
				})
			}

			recordSet := &route53.ResourceRecordSet{
				Name:            &recordName,
				Type:            &recordSetType,
				TTL:             &recordSetTtl,
				ResourceRecords: records,
			}

			change := &route53.Change{
				Action:            &action,
				ResourceRecordSet: recordSet,
			}

			changes = append(changes, change)
		} else if addr != "" {
			action := "UPSERT"
			recordSetType := "A"
			recordSetTtl := int64(60)
			records := []*route53.ResourceRecord{
				&route53.ResourceRecord{
					Value: &addr,
				},
			}

			recordSet := &route53.ResourceRecordSet{
				Name:            &recordName,
				Type:            &recordSetType,
				TTL:             &recordSetTtl,
				ResourceRecords: records,
			}

			change := &route53.Change{
				Action:            &action,
				ResourceRecordSet: recordSet,
			}

			changes = append(changes, change)
		}
	}

	if len(curAddrs6) != 1 || curAddrs6[0] != addr6 {
		if addr6 == "" && len(curAddrs6) > 0 {
			action := "DELETE"
			recordSetType := "AAAA"
			recordSetTtl := int64(60)
			records := []*route53.ResourceRecord{}

			for _, a6 := range curAddrs6 {
				records = append(records, &route53.ResourceRecord{
					Value: &a6,
				})
			}

			recordSet := &route53.ResourceRecordSet{
				Name:            &recordName,
				Type:            &recordSetType,
				TTL:             &recordSetTtl,
				ResourceRecords: records,
			}

			change := &route53.Change{
				Action:            &action,
				ResourceRecordSet: recordSet,
			}

			changes = append(changes, change)
		} else if addr6 != "" {
			action := "UPSERT"
			recordSetType := "AAAA"
			recordSetTtl := int64(60)
			records := []*route53.ResourceRecord{
				&route53.ResourceRecord{
					Value: &addr6,
				},
			}

			recordSet := &route53.ResourceRecordSet{
				Name:            &recordName,
				Type:            &recordSetType,
				TTL:             &recordSetTtl,
				ResourceRecords: records,
			}

			change := &route53.Change{
				Action:            &action,
				ResourceRecordSet: recordSet,
			}

			changes = append(changes, change)
		}
	}

	if len(changes) > 0 {
		_, err = servc.ChangeResourceRecordSets(
			&route53.ChangeResourceRecordSetsInput{
				HostedZoneId: &zoneId,
				ChangeBatch: &route53.ChangeBatch{
					Changes: changes,
				},
			},
		)
		if err != nil {
			err = &errortypes.RequestError{
				errors.Wrap(err, "domain: Failed to list Route53 records"),
			}
			return
		}
	}

	return
}
