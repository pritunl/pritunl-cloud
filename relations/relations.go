package relations

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
)

type Query struct {
	Id         any
	Label      string
	Collection string
	Project    []Project
	Relations  []Relation
}

type Relation struct {
	Key          string
	Label        string
	From         string
	LocalField   string
	ForeignField string
	Sort         map[string]int
	Project      []Project
	Relations    []Relation
}

type Project struct {
	Key    string
	Keys   []string
	Label  string
	Format func(values ...any) any
}

func (r *Query) addRelation(pipeline []bson.M, relation Relation) []bson.M {
	lookup := bson.M{
		"from":         relation.From,
		"localField":   relation.LocalField,
		"foreignField": relation.ForeignField,
		"as":           relation.From,
	}

	if len(relation.Project) > 0 || len(relation.Sort) > 0 ||
		len(relation.Relations) > 0 {

		nestedPipeline := []bson.M{}
		if len(relation.Project) > 0 {
			projection := bson.M{
				"_id": 1,
			}
			for _, proj := range relation.Project {
				if len(proj.Keys) > 0 {
					for _, key := range proj.Keys {
						projection[key] = 1
					}
				} else {
					projection[proj.Key] = 1
				}
			}
			nestedPipeline = append(nestedPipeline, bson.M{
				"$project": projection,
			})
		}

		if len(relation.Sort) > 0 {
			nestedPipeline = append(nestedPipeline, bson.M{
				"$sort": relation.Sort,
			})
		}

		for _, nestedRelation := range relation.Relations {
			nestedPipeline = r.addRelation(nestedPipeline, nestedRelation)
		}

		lookup["pipeline"] = nestedPipeline
	}

	return append(pipeline, bson.M{
		"$lookup": lookup,
	})
}

func (r *Query) convertToResponse(doc bson.M) *Response {
	response := &Response{
		Id:        doc["_id"],
		Label:     r.Label,
		Fields:    []Field{},
		Relations: []Related{},
	}

	for _, proj := range r.Project {
		if len(proj.Keys) > 0 {
			value, ok := doc[proj.Keys[0]]
			if ok {
				if proj.Format != nil {
					values := []any{}

					for _, key := range proj.Keys {
						val, ok := doc[key]
						if !ok {
							values = append(values, nil)
						} else {
							values = append(values, val)
						}
					}

					value = proj.Format(values...)
				}

				response.Fields = append(response.Fields, Field{
					Key:   proj.Key,
					Label: proj.Label,
					Value: value,
				})
			}
		} else {
			value, ok := doc[proj.Key]
			if ok {
				if proj.Format != nil {
					value = proj.Format(value)
				}

				response.Fields = append(response.Fields, Field{
					Key:   proj.Key,
					Label: proj.Label,
					Value: value,
				})
			}
		}
	}

	for _, relation := range r.Relations {
		docs, ok := doc[relation.From].(primitive.A)
		if ok {
			response.Relations = append(
				response.Relations,
				r.convertToRelated(relation, docs),
			)
		}
	}

	return response
}

func (r *Query) convertToRelated(relation Relation,
	docs primitive.A) Related {

	related := Related{
		Label:     relation.Label,
		Resources: []Resource{},
	}

	for _, docInf := range docs {
		doc, ok := docInf.(bson.M)
		if !ok {
			continue
		}

		resource := Resource{
			Id:        doc["_id"],
			Type:      relation.Label,
			Fields:    []Field{},
			Relations: []Related{},
		}

		for _, proj := range relation.Project {
			if len(proj.Keys) > 0 {
				value, ok := doc[proj.Keys[0]]
				if ok {
					if proj.Format != nil {
						values := []any{}

						for _, key := range proj.Keys {
							val, ok := doc[key]
							if !ok {
								values = append(values, nil)
							} else {
								values = append(values, val)
							}
						}

						value = proj.Format(values...)
					}

					resource.Fields = append(resource.Fields, Field{
						Key:   proj.Key,
						Label: proj.Label,
						Value: value,
					})
				}
			} else {
				value, ok := doc[proj.Key]
				if ok {
					if proj.Format != nil {
						value = proj.Format(value)
					}

					resource.Fields = append(resource.Fields, Field{
						Key:   proj.Key,
						Label: proj.Label,
						Value: value,
					})
				}
			}
		}

		for _, relation := range relation.Relations {
			docs, ok := doc[relation.From].(primitive.A)
			if ok {
				resource.Relations = append(
					resource.Relations,
					r.convertToRelated(relation, docs),
				)
			}
		}

		related.Resources = append(related.Resources, resource)
	}

	return related
}

func (r *Query) Aggregate(db *database.Database) (
	resp *Response, err error) {

	coll := db.GetCollection(r.Collection)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_id": r.Id,
			},
		},
	}

	if len(r.Project) > 0 {
		projection := bson.M{"_id": 1}
		for _, proj := range r.Project {
			if len(proj.Keys) > 0 {
				for _, key := range proj.Keys {
					projection[key] = 1
				}
			} else {
				projection[proj.Key] = 1
			}
		}
		pipeline = append(pipeline, bson.M{"$project": projection})
	}

	for _, relation := range r.Relations {
		pipeline = r.addRelation(pipeline, relation)
	}

	cursor, err := coll.Aggregate(db, pipeline)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(db)

	var results []bson.M
	err = cursor.All(db, &results)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	if len(results) == 0 {
		err = &database.NotFoundError{
			errors.New("relations: Resource not found"),
		}
		return
	}

	resp = r.convertToResponse(results[0])
	return
}
