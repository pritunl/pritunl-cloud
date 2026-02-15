package advisory

import (
	"net/http"
	"time"

	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

type Advisory struct {
	Id              string    `bson:"id" json:"id"`
	Timestamp       time.Time `bson:"timestamp" json:"timestamp"`
	Status          string    `bson:"status" json:"status"`
	Description     string    `bson:"description" json:"description"`
	Score           float64   `bson:"score" json:"score"`
	Severity        string    `bson:"severity" json:"severity"`
	Vector          string    `bson:"vector" json:"vector"`
	Complexity      string    `bson:"complexity" json:"complexity"`
	Privileges      string    `bson:"privileges" json:"privileges"`
	Interaction     string    `bson:"interaction" json:"interaction"`
	Scope           string    `bson:"scope" json:"scope"`
	Confidentiality string    `bson:"confidentiality" json:"confidentiality"`
	Integrity       string    `bson:"integrity" json:"integrity"`
	Availability    string    `bson:"availability" json:"availability"`
}

func (a *Advisory) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if a.Id == "" {
		errData = &errortypes.ErrorData{
			Error:   "id_required",
			Message: "Missing required advisory ID",
		}
		return
	}

	if !cveIdReg.MatchString(a.Id) {
		errData = &errortypes.ErrorData{
			Error:   "id_invalid",
			Message: "Invalid advisory ID",
		}
		return
	}

	if a.Status != "" && !ValidStatuses.Contains(a.Status) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_status",
			Message: "Invalid advisory status",
		}
		return
	}

	if a.Score < 0 || a.Score > 10 {
		errData = &errortypes.ErrorData{
			Error:   "invalid_score",
			Message: "Advisory score must be between 0 and 10",
		}
		return
	}

	if a.Severity != "" && !ValidSeverities.Contains(a.Severity) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_severity",
			Message: "Invalid advisory severity",
		}
		return
	}

	if a.Vector != "" && !ValidVectors.Contains(a.Vector) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_vector",
			Message: "Invalid advisory attack vector",
		}
		return
	}

	if a.Complexity != "" && !ValidComplexities.Contains(a.Complexity) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_complexity",
			Message: "Invalid advisory attack complexity",
		}
		return
	}

	if a.Privileges != "" && !ValidPrivileges.Contains(a.Privileges) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_privileges",
			Message: "Invalid advisory privileges required",
		}
		return
	}

	if a.Interaction != "" && !ValidInteractions.Contains(a.Interaction) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_interaction",
			Message: "Invalid advisory user interaction",
		}
		return
	}

	if a.Scope != "" && !ValidScopes.Contains(a.Scope) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_scope",
			Message: "Invalid advisory scope",
		}
		return
	}

	if a.Confidentiality != "" &&
		!ValidImpacts.Contains(a.Confidentiality) {

		errData = &errortypes.ErrorData{
			Error:   "invalid_confidentiality",
			Message: "Invalid advisory confidentiality impact",
		}
		return
	}

	if a.Integrity != "" && !ValidImpacts.Contains(a.Integrity) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_integrity",
			Message: "Invalid advisory integrity impact",
		}
		return
	}

	if a.Availability != "" && !ValidImpacts.Contains(a.Availability) {
		errData = &errortypes.ErrorData{
			Error:   "invalid_availability",
			Message: "Invalid advisory availability impact",
		}
		return
	}

	return
}
