package advisory

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type nvdResponse struct {
	TotalResults    int `json:"totalResults"`
	Vulnerabilities []struct {
		Cve struct {
			ID           string `json:"id"`
			VulnStatus   string `json:"vulnStatus"`
			Descriptions []struct {
				Lang  string `json:"lang"`
				Value string `json:"value"`
			} `json:"descriptions"`
			Metrics struct {
				CvssMetricV31 []struct {
					Type     string `json:"type"`
					CvssData struct {
						VectorString          string  `json:"vectorString"`
						BaseScore             float64 `json:"baseScore"`
						BaseSeverity          string  `json:"baseSeverity"`
						AttackVector          string  `json:"attackVector"`
						AttackComplexity      string  `json:"attackComplexity"`
						PrivilegesRequired    string  `json:"privilegesRequired"`
						UserInteraction       string  `json:"userInteraction"`
						Scope                 string  `json:"scope"`
						ConfidentialityImpact string  `json:"confidentialityImpact"`
						IntegrityImpact       string  `json:"integrityImpact"`
						AvailabilityImpact    string  `json:"availabilityImpact"`
					} `json:"cvssData"`
				} `json:"cvssMetricV31"`
			} `json:"metrics"`
		} `json:"cve"`
	} `json:"vulnerabilities"`
}

func normalizeStatus(status string) string {
	switch status {
	case "Analyzed":
		return Analyzed
	case "Awaiting Analysis":
		return AwaitingAnalysis
	case "Rejected":
		return Rejected
	case "Undergoing Analysis":
		return Undergoing
	case "Modified":
		return Modified
	case "Deferred":
		return Deferred
	default:
		return strings.ToLower(strings.ReplaceAll(status, " ", "_"))
	}
}

func normalizeValue(val string) string {
	switch strings.ToUpper(val) {
	case "NONE":
		return None
	case "LOW":
		return Low
	case "MEDIUM":
		return Medium
	case "HIGH":
		return High
	case "CRITICAL":
		return Critical
	case "NETWORK":
		return Network
	case "ADJACENT_NETWORK", "ADJACENT":
		return Adjacent
	case "LOCAL":
		return Local
	case "PHYSICAL":
		return Physical
	case "REQUIRED":
		return Required
	case "UNCHANGED":
		return Unchanged
	case "CHANGED":
		return Changed
	default:
		return strings.ToLower(val)
	}
}

func GetOne(db *database.Database, query *bson.M) (adv *Advisory, err error) {
	coll := db.Advisories()
	adv = &Advisory{}

	err = coll.FindOne(db, query).Decode(adv)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func Fetch(db *database.Database, cveId string) (adv *Advisory, err error) {
	adv, err = GetOne(db, &bson.M{
		"_id": cveId,
	})
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			adv = nil
		} else {
			return
		}
	}

	if adv != nil && time.Since(adv.Timestamp) < 24*time.Hour {
		return
	}

	req, err := http.NewRequest("GET", nvdApi, nil)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "advisory: Failed to create request"),
		}
		return
	}

	query := req.URL.Query()
	query.Set("cveId", cveId)
	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "advisory: Request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = &errortypes.RequestError{
			errors.Newf("advisory: Bad status status %d", resp.StatusCode),
		}
		return
	}

	nvdResp := &nvdResponse{}
	err = json.NewDecoder(resp.Body).Decode(nvdResp)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "advisory: Failed to decode response"),
		}
		return
	}

	if nvdResp.TotalResults == 0 || len(nvdResp.Vulnerabilities) == 0 {
		err = &errortypes.NotFoundError{
			errors.New("advisory: Not found"),
		}
		return
	}

	cve := nvdResp.Vulnerabilities[0].Cve

	adv = &Advisory{
		Status: normalizeStatus(cve.VulnStatus),
	}

	for _, desc := range cve.Descriptions {
		if desc.Lang == "en" {
			adv.Description = desc.Value
			break
		}
	}

	metrics := cve.Metrics.CvssMetricV31
	if len(metrics) > 0 {
		var cvss *struct {
			Type     string `json:"type"`
			CvssData struct {
				VectorString          string  `json:"vectorString"`
				BaseScore             float64 `json:"baseScore"`
				BaseSeverity          string  `json:"baseSeverity"`
				AttackVector          string  `json:"attackVector"`
				AttackComplexity      string  `json:"attackComplexity"`
				PrivilegesRequired    string  `json:"privilegesRequired"`
				UserInteraction       string  `json:"userInteraction"`
				Scope                 string  `json:"scope"`
				ConfidentialityImpact string  `json:"confidentialityImpact"`
				IntegrityImpact       string  `json:"integrityImpact"`
				AvailabilityImpact    string  `json:"availabilityImpact"`
			} `json:"cvssData"`
		}

		for i := range metrics {
			if metrics[i].Type == "Primary" {
				cvss = &metrics[i]
				break
			}
		}
		if cvss == nil {
			cvss = &metrics[0]
		}

		adv.Id = strings.ToUpper(cve.ID)
		adv.Timestamp = time.Now()
		adv.Score = cvss.CvssData.BaseScore
		adv.Severity = normalizeValue(cvss.CvssData.BaseSeverity)
		adv.Vector = normalizeValue(cvss.CvssData.AttackVector)
		adv.Complexity = normalizeValue(cvss.CvssData.AttackComplexity)
		adv.Privileges = normalizeValue(cvss.CvssData.PrivilegesRequired)
		adv.Interaction = normalizeValue(cvss.CvssData.UserInteraction)
		adv.Scope = normalizeValue(cvss.CvssData.Scope)
		adv.Confidentiality = normalizeValue(
			cvss.CvssData.ConfidentialityImpact)
		adv.Integrity = normalizeValue(cvss.CvssData.IntegrityImpact)
		adv.Availability = normalizeValue(cvss.CvssData.AvailabilityImpact)
	}

	errData, err := adv.Validate(db)
	if err != nil {
		return
	}
	if errData != nil {
		err = errData.GetError()
		return
	}

	err = adv.Commit(db)
	if err != nil {
		return
	}

	return
}
