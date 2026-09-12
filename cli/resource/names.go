package resource

import (
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/cli/widget"
	"github.com/pritunl/pritunl-cloud/database"
)

// IdSet collects distinct non zero ids referenced by a page of items.
type IdSet map[bson.ObjectID]struct{}

func (s IdSet) Add(id bson.ObjectID) {
	if !id.IsZero() {
		s[id] = struct{}{}
	}
}

func (s IdSet) List() []bson.ObjectID {
	ids := make([]bson.ObjectID, 0, len(s))
	for id := range s {
		ids = append(ids, id)
	}
	return ids
}

// Names returns the names of the documents with the ids in the
// collection, used to show names in place of ids on cards.
func Names(db *database.Database, coll *database.Collection,
	ids []bson.ObjectID) (names map[bson.ObjectID]string, err error) {

	names = map[bson.ObjectID]string{}
	if len(ids) == 0 {
		return
	}

	cursor, err := coll.Find(
		db,
		bson.M{
			"_id": bson.M{
				"$in": ids,
			},
		},
		options.Find().SetProjection(bson.D{
			{"name", 1},
		}),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		named := &database.Named{}
		err = cursor.Decode(named)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		names[named.Id] = named.Name
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

// AllNames returns the ids and names of the documents matching the query
// sorted by name, used for filter select options.
func AllNames(db *database.Database, coll *database.Collection,
	query bson.M) (items []*database.Named, err error) {

	items = []*database.Named{}

	cursor, err := coll.Find(
		db,
		query,
		options.Find().
			SetSort(bson.D{
				{"name", 1},
			}).
			SetProjection(bson.D{
				{"name", 1},
			}),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		named := &database.Named{}
		err = cursor.Decode(named)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		items = append(items, named)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

// SelectOptions converts named documents to filter select options with
// the id hex as the value.
func SelectOptions(items []*database.Named) []widget.SelectOption {
	opts := make([]widget.SelectOption, 0, len(items))
	for _, item := range items {
		opts = append(opts, widget.SelectOption{
			Label: item.Name,
			Value: item.Id.Hex(),
		})
	}
	return opts
}

// OrgNamed is a document name with its organization, used for select
// options filtered by the organization of another input such as the
// secrets and certificates of an organization.
type OrgNamed struct {
	Id           bson.ObjectID `bson:"_id"`
	Name         string        `bson:"name"`
	Organization bson.ObjectID `bson:"organization"`
}

// AllOrgNames returns the ids, names and organizations of the documents
// matching the query sorted by name.
func AllOrgNames(db *database.Database, coll *database.Collection,
	query bson.M) (items []*OrgNamed, err error) {

	items = []*OrgNamed{}

	cursor, err := coll.Find(
		db,
		query,
		options.Find().
			SetSort(bson.D{
				{"name", 1},
			}).
			SetProjection(bson.D{
				{"name", 1},
				{"organization", 1},
			}),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		item := &OrgNamed{}
		err = cursor.Decode(item)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		items = append(items, item)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

// OrgSelectOptions returns the select options of the documents in the
// organization, all documents when the organization is zero.
func OrgSelectOptions(items []*OrgNamed,
	orgId bson.ObjectID) []widget.SelectOption {

	opts := []widget.SelectOption{}
	for _, item := range items {
		if item.Organization != orgId {
			continue
		}
		opts = append(opts, widget.SelectOption{
			Label: item.Name,
			Value: item.Id.Hex(),
		})
	}
	return opts
}

// NameMap indexes named documents by id.
func NameMap(items []*database.Named) map[bson.ObjectID]string {
	names := map[bson.ObjectID]string{}
	for _, item := range items {
		names[item.Id] = item.Name
	}
	return names
}

// ParseId returns the object id of a hex string or the zero id.
func ParseId(hex string) bson.ObjectID {
	id, _ := bson.ObjectIDFromHex(hex)
	return id
}

// FormatTime formats a timestamp in local time or returns an empty
// string for the zero time.
func FormatTime(timestamp time.Time) string {
	if timestamp.IsZero() {
		return ""
	}
	return timestamp.Local().Format("2006-01-02 15:04:05")
}

// IdHex returns the hex of an id or an empty string for the zero id.
func IdHex(id bson.ObjectID) string {
	if id.IsZero() {
		return ""
	}
	return id.Hex()
}

// WithSelect prepends a placeholder choice such as Select Zone to the
// options.
func WithSelect(label string, opts []widget.SelectOption) []widget.SelectOption {
	return append([]widget.SelectOption{{Label: label, Value: ""}}, opts...)
}
