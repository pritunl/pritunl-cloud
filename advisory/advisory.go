package advisory

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	None     = "none"
	Low      = "low"
	Medium   = "medium"
	High     = "high"
	Critical = "critical"
	Network  = "network"
	Adjacent = "adjacent"
	Local    = "local"
	Physical = "physical"
	Required = "required"

	Unchanged = "unchanged"
	Changed   = "changed"

	Analyzed         = "analyzed"
	AwaitingAnalysis = "awaiting_analysis"
	Rejected         = "rejected"
	Undergoing       = "undergoing_analysis"
	Modified         = "modified"
	Deferred         = "deferred"

	nvdApi = "https://services.nvd.nist.gov/rest/json/cves/2.0"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

type Advisory struct {
	Status          string  `bson:"status" json:"status"`
	Description     string  `bson:"description" json:"description"`
	Score           float64 `bson:"score" json:"score"`
	Severity        string  `bson:"severity" json:"severity"`
	Vector          string  `bson:"vector" json:"vector"`
	Complexity      string  `bson:"complexity" json:"complexity"`
	Privileges      string  `bson:"privileges" json:"privileges"`
	Interaction     string  `bson:"interaction" json:"interaction"`
	Scope           string  `bson:"scope" json:"scope"`
	Confidentiality string  `bson:"confidentiality" json:"confidentiality"`
	Integrity       string  `bson:"integrity" json:"integrity"`
	Availability    string  `bson:"availability" json:"availability"`
}

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

func Fetch(cveId string) (adv *Advisory, err error) {
	req, err := http.NewRequest("GET", nvdApi, nil)
	if err != nil {
		err = fmt.Errorf("advisory: Failed to create request %w", err)
		return
	}

	query := req.URL.Query()
	query.Set("cveId", cveId)
	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("advisory: Request failed %w", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("advisory: NVD API returned status %d", resp.StatusCode)
		return
	}

	nvdResp := &nvdResponse{}
	err = json.NewDecoder(resp.Body).Decode(nvdResp)
	if err != nil {
		err = fmt.Errorf("advisory: Failed to decode response %w", err)
		return
	}

	if nvdResp.TotalResults == 0 || len(nvdResp.Vulnerabilities) == 0 {
		err = fmt.Errorf("advisory: CVE %s not found", cveId)
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

		adv.Score = cvss.CvssData.BaseScore
		adv.Severity = normalizeValue(cvss.CvssData.BaseSeverity)
		adv.Vector = normalizeValue(cvss.CvssData.AttackVector)
		adv.Complexity = normalizeValue(cvss.CvssData.AttackComplexity)
		adv.Privileges = normalizeValue(cvss.CvssData.PrivilegesRequired)
		adv.Interaction = normalizeValue(cvss.CvssData.UserInteraction)
		adv.Scope = normalizeValue(cvss.CvssData.Scope)
		adv.Confidentiality = normalizeValue(cvss.CvssData.ConfidentialityImpact)
		adv.Integrity = normalizeValue(cvss.CvssData.IntegrityImpact)
		adv.Availability = normalizeValue(cvss.CvssData.AvailabilityImpact)
	}

	return
}
