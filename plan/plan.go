package plan

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/eval"
	"github.com/pritunl/pritunl-cloud/utils"
)

type Plan struct {
	Id           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	Comment      string        `bson:"comment" json:"comment"`
	Organization bson.ObjectID `bson:"organization" json:"organization"`
	Statements   []*Statement  `bson:"statements" json:"statements"`
}

type Completion struct {
	Id           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	Organization bson.ObjectID `bson:"organization" json:"organization"`
}

type Statement struct {
	Id        bson.ObjectID `bson:"id" json:"id"`
	Statement string        `bson:"statement" json:"statement"`
}

func (p *Plan) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	p.Name = utils.FilterName(p.Name)

	if p.Organization.IsZero() {
		errData = &errortypes.ErrorData{
			Error:   "organization_required",
			Message: "Missing required organization",
		}
		return
	}

	emptyData, err := GetEmtpyData()
	if err != nil {
		return
	}

	if p.Statements == nil {
		p.Statements = []*Statement{}
	}

	for _, statement := range p.Statements {
		if statement.Id.IsZero() {
			statement.Id = bson.NewObjectID()
		}

		err = eval.Validate(statement.Statement)
		if err != nil {
			return
		}

		_, _, err = eval.Eval(emptyData, statement.Statement)
		if err != nil {
			return
		}
	}

	return
}

func (p *Plan) UpdateStatements(inStatements []*Statement) (err error) {
	curStatements := map[bson.ObjectID]*Statement{}
	for _, statement := range p.Statements {
		curStatements[statement.Id] = statement
	}

	newStatements := []*Statement{}
	for _, statement := range inStatements {
		curStatement := curStatements[statement.Id]
		if curStatement != nil {
			if statement.Statement == curStatement.Statement {
				newStatements = append(newStatements, curStatement)
			} else {
				newStatement := &Statement{
					Id:        bson.NewObjectID(),
					Statement: statement.Statement,
				}
				newStatements = append(newStatements, newStatement)
			}
		} else {
			newStatement := &Statement{
				Id:        bson.NewObjectID(),
				Statement: statement.Statement,
			}
			newStatements = append(newStatements, newStatement)
		}
	}

	p.Statements = newStatements

	return
}

func (p *Plan) Commit(db *database.Database) (err error) {
	coll := db.Plans()

	err = coll.Commit(p.Id, p)
	if err != nil {
		return
	}

	return
}

func (p *Plan) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Plans()

	err = coll.CommitFields(p.Id, p, fields)
	if err != nil {
		return
	}

	return
}

func (p *Plan) Insert(db *database.Database) (err error) {
	coll := db.Plans()

	if !p.Id.IsZero() {
		err = &errortypes.DatabaseError{
			errors.New("domain: Plan already exists"),
		}
		return
	}

	_, err = coll.InsertOne(db, p)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}
