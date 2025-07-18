package relations

import (
	"fmt"
	"strings"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
)

var registry = map[string]Query{}

func Register(kind string, definition Query) {
	registry[kind] = definition
}

func Aggregate(db *database.Database, kind string, id primitive.ObjectID) (
	resp *Response, err error) {

	definition, ok := registry[kind]
	if !ok {
		return
	}

	definition.Id = id

	resp, err = definition.Aggregate(db)
	if err != nil {
		return
	}

	return
}

func AggregateOrg(db *database.Database, kind string,
	orgId, id primitive.ObjectID) (resp *Response, err error) {

	definition, ok := registry[kind]
	if !ok {
		return
	}

	definition.Id = id
	definition.Organization = orgId

	resp, err = definition.Aggregate(db)
	if err != nil {
		return
	}

	return
}

func blockDelete(resources []Resource) string {
	for _, resource := range resources {
		if resource.BlockDelete {
			return resource.Type
		}

		for _, related := range resource.Relations {
			label := blockDelete(related.Resources)
			if label != "" {
				return label
			}
		}
	}
	return ""
}

func CanDelete(db *database.Database, kind string, id primitive.ObjectID) (
	errData *errortypes.ErrorData, err error) {

	definition, ok := registry[kind]
	if !ok {
		return
	}

	definition.Id = id

	resp, err := definition.Aggregate(db)
	if err != nil {
		return
	}

	if resp.DeleteProtection {
		errData = &errortypes.ErrorData{
			Error:   "delete_protected_resource",
			Message: "Cannot delete resource with delete protection enabled",
		}
		return
	}

	labels := []string{}
	for _, related := range resp.Relations {
		label := blockDelete(related.Resources)
		if label != "" {
			labels = append(labels, label)
		}
	}

	if len(labels) > 0 {
		errData = &errortypes.ErrorData{
			Error: "related_resources_exist",
			Message: fmt.Sprintf(
				"Related [%s] resources must be deleted first. "+
					"Check resource overview",
				strings.Join(labels, ", "),
			),
		}
		return
	}

	return
}

func CanDeleteOrg(db *database.Database, kind string,
	orgId, id primitive.ObjectID) (errData *errortypes.ErrorData, err error) {

	definition, ok := registry[kind]
	if !ok {
		return
	}

	definition.Id = id
	definition.Organization = orgId

	resp, err := definition.Aggregate(db)
	if err != nil {
		return
	}

	if resp.DeleteProtection {
		errData = &errortypes.ErrorData{
			Error:   "delete_protected_resource",
			Message: "Cannot delete resource with delete protection enabled",
		}
		return
	}

	labels := []string{}
	for _, related := range resp.Relations {
		label := blockDelete(related.Resources)
		if label != "" {
			labels = append(labels, label)
		}
	}

	if len(labels) > 0 {
		errData = &errortypes.ErrorData{
			Error: "related_resources_exist",
			Message: fmt.Sprintf(
				"Related [%s] resources must be deleted first. "+
					"Check resource overview",
				strings.Join(labels, ", "),
			),
		}
		return
	}

	return
}

func CanDeleteAll(db *database.Database, kind string,
	ids []primitive.ObjectID) (errData *errortypes.ErrorData, err error) {

	for _, id := range ids {
		errData, err = CanDelete(db, kind, id)
		if err != nil {
			return
		}

		if errData != nil {
			return
		}
	}

	return
}

func CanDeleteOrgAll(db *database.Database, kind string,
	orgId primitive.ObjectID, ids []primitive.ObjectID) (
	errData *errortypes.ErrorData, err error) {

	for _, id := range ids {
		errData, err = CanDeleteOrg(db, kind, orgId, id)
		if err != nil {
			return
		}

		if errData != nil {
			return
		}
	}

	return
}
