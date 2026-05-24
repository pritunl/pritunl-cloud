package advisory

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
)

var (
	lastCall time.Time
)

type nvdCvssData struct {
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
}

type nvdCvssMetric struct {
	Type     string      `json:"type"`
	CvssData nvdCvssData `json:"cvssData"`
}

type nvdMetrics struct {
	CvssMetricV31 []nvdCvssMetric `json:"cvssMetricV31"`
}

type nvdDescription struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

type nvdCve struct {
	ID           string           `json:"id"`
	VulnStatus   string           `json:"vulnStatus"`
	Descriptions []nvdDescription `json:"descriptions"`
	Metrics      nvdMetrics       `json:"metrics"`
}

type nvdVulnerability struct {
	Cve nvdCve `json:"cve"`
}

type nvdResponse struct {
	TotalResults    int                `json:"totalResults"`
	Vulnerabilities []nvdVulnerability `json:"vulnerabilities"`
}

type redhatCvss3 struct {
	BaseScore     string `json:"cvss3_base_score"`
	ScoringVector string `json:"cvss3_scoring_vector"`
	Status        string `json:"status"`
}

type redhatResponse struct {
	Name           string      `json:"name"`
	ThreatSeverity string      `json:"threat_severity"`
	PublicDate     string      `json:"public_date"`
	Details        []string    `json:"details"`
	Cvss3          redhatCvss3 `json:"cvss3"`
}

func normalizeRedhatSeverity(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return Critical
	case "important":
		return High
	case "moderate":
		return Medium
	case "low":
		return Low
	case "none":
		return None
	default:
		return strings.ToLower(severity)
	}
}

func parseCvss3Vector(vector string) (av, ac, pr, ui, scope, c, i, a string) {
	parts := strings.Split(vector, "/")
	for _, part := range parts {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "AV":
			switch kv[1] {
			case "N":
				av = Network
			case "A":
				av = Adjacent
			case "L":
				av = Local
			case "P":
				av = Physical
			}
		case "AC":
			switch kv[1] {
			case "L":
				ac = Low
			case "H":
				ac = High
			}
		case "PR":
			switch kv[1] {
			case "N":
				pr = None
			case "L":
				pr = Low
			case "H":
				pr = High
			}
		case "UI":
			switch kv[1] {
			case "N":
				ui = None
			case "R":
				ui = Required
			}
		case "S":
			switch kv[1] {
			case "U":
				scope = Unchanged
			case "C":
				scope = Changed
			}
		case "C":
			c = impactFromCode(kv[1])
		case "I":
			i = impactFromCode(kv[1])
		case "A":
			a = impactFromCode(kv[1])
		}
	}
	return
}

func impactFromCode(code string) string {
	switch code {
	case "N":
		return None
	case "L":
		return Low
	case "H":
		return High
	}
	return ""
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

func getIdPrefix() string {
	switch settings.Telemetry.CveSource {
	case RedHat:
		return "rh:"
	case Nist:
		return "nvd:"
	}

	return ""
}

func getOne(db *database.Database, query *bson.M) (adv *Advisory, err error) {
	coll := db.Advisories()
	adv = &Advisory{}

	err = coll.FindOne(db, query).Decode(adv)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func getOneNvd(db *database.Database, cveId string) (
	adv *Advisory, err error) {

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

	nvdApiKey := settings.Telemetry.NvdApiKey
	if nvdApiKey != "" {
		req.Header.Set("apiKey", nvdApiKey)
	}

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
		Id:        "nvd:" + strings.ToUpper(cve.ID),
		Timestamp: time.Now(),
		Status:    normalizeStatus(cve.VulnStatus),
	}

	for _, desc := range cve.Descriptions {
		if desc.Lang == "en" {
			adv.Description = desc.Value
			break
		}
	}

	metrics := cve.Metrics.CvssMetricV31
	if len(metrics) > 0 {
		var cvss *nvdCvssMetric

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

func getOneRedhat(db *database.Database, cveId string) (
	adv *Advisory, err error) {

	u, err := url.Parse(redhatApi)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "advisory: Failed to parse redhat url"),
		}
		return
	}

	u.Path = fmt.Sprintf(redhatApiPath, strings.ToUpper(cveId))

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "advisory: Failed to create request"),
		}
		return
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "advisory: Request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		err = &errortypes.NotFoundError{
			errors.New("advisory: Not found"),
		}
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = &errortypes.RequestError{
			errors.Newf("advisory: Bad status status %d", resp.StatusCode),
		}
		return
	}

	rhResp := &redhatResponse{}
	err = json.NewDecoder(resp.Body).Decode(rhResp)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "advisory: Failed to decode response"),
		}
		return
	}

	if rhResp.Name == "" {
		err = &errortypes.NotFoundError{
			errors.New("advisory: Not found"),
		}
		return
	}

	adv = &Advisory{
		Id:        "rh:" + strings.ToUpper(rhResp.Name),
		Timestamp: time.Now(),
		Status:    Analyzed,
		Severity:  normalizeRedhatSeverity(rhResp.ThreatSeverity),
	}

	if len(rhResp.Details) > 0 {
		adv.Description = strings.Join(rhResp.Details, "\n\n")
	}

	if rhResp.Cvss3.BaseScore != "" {
		score, e := strconv.ParseFloat(rhResp.Cvss3.BaseScore, 64)
		if e == nil {
			adv.Score = score
		}
	}

	if rhResp.Cvss3.ScoringVector != "" {
		av, ac, pr, ui, scope, c, i, a := parseCvss3Vector(
			rhResp.Cvss3.ScoringVector)
		adv.Vector = av
		adv.Complexity = ac
		adv.Privileges = pr
		adv.Interaction = ui
		adv.Scope = scope
		adv.Confidentiality = c
		adv.Integrity = i
		adv.Availability = a
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

func GetOneLimit(db *database.Database, cveId string) (
	adv *Advisory, err error) {

	adv, err = getOne(db, &bson.M{
		"_id": getIdPrefix() + cveId,
	})
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			adv = nil
		} else {
			return
		}
	}

	if adv.IsFresh() {
		return
	}

	since := time.Since(lastCall)
	var limit time.Duration
	if settings.Telemetry.CveSource == RedHat {
		limit = time.Duration(
			settings.Telemetry.RedhatApiLimit) * time.Second
	} else if settings.Telemetry.NvdApiKey != "" {
		limit = time.Duration(
			settings.Telemetry.NvdApiAuthLimit) * time.Second
	} else {
		limit = time.Duration(settings.Telemetry.NvdApiLimit) * time.Second
	}
	if since < limit {
		time.Sleep(limit - since)
	}
	lastCall = time.Now()

	if settings.Telemetry.CveSource == RedHat {
		adv, err = getOneRedhat(db, cveId)
	} else {
		adv, err = getOneNvd(db, cveId)
	}
	if err != nil {
		return
	}

	return
}

func GetOne(db *database.Database, cveId string) (adv *Advisory, err error) {
	adv, err = getOne(db, &bson.M{
		"_id": getIdPrefix() + cveId,
	})
	if err != nil {
		if _, ok := err.(*database.NotFoundError); ok {
			adv = nil
		} else {
			return
		}
	}

	if adv.IsFresh() {
		return
	}

	if settings.Telemetry.CveSource == RedHat {
		adv, err = getOneRedhat(db, cveId)
	} else {
		adv, err = getOneNvd(db, cveId)
	}
	if err != nil {
		return
	}

	return
}

func GetOneForce(db *database.Database, cveId string) (
	adv *Advisory, err error) {

	if settings.Telemetry.CveSource == RedHat {
		adv, err = getOneRedhat(db, cveId)
	} else {
		adv, err = getOneNvd(db, cveId)
	}
	if err != nil {
		return
	}

	return
}

func Remove(db *database.Database, cveId string) (err error) {
	coll := db.Advisories()

	_, err = coll.DeleteOne(db, &bson.M{
		"_id": getIdPrefix() + cveId,
	})
	if err != nil {
		err = database.ParseError(err)
		if _, ok := err.(*database.NotFoundError); ok {
			err = nil
		} else {
			return
		}
	}

	return
}
