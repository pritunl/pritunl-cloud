package nodeport

import (
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/tools/set"
)

type PortRange struct {
	Start int
	End   int
}

func (r *PortRange) Contains(port int) bool {
	if port >= r.Start && port <= r.End {
		return true
	}
	return false
}

func Get(db *database.Database, resourceId, ndePrtId primitive.ObjectID) (
	ndePrt *NodePort, err error) {

	coll := db.NodePorts()
	ndePrt = &NodePort{}

	err = coll.FindOne(db, &bson.M{
		"_id":      ndePrtId,
		"resource": resourceId,
	}).Decode(ndePrt)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	return
}

func GetPortRanges() (ranges []*PortRange, err error) {
	ranges = []*PortRange{}
	parts := strings.Split(settings.Hypervisor.NodePortRanges, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		bounds := strings.Split(part, "-")
		if len(bounds) != 2 {
			err = &errortypes.ParseError{
				errors.New("nodeport: Invalid port range format"),
			}
			return
		}

		start, e := strconv.Atoi(strings.TrimSpace(bounds[0]))
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "nodeport: Invalid start port"),
			}
			return
		}

		end, e := strconv.Atoi(strings.TrimSpace(bounds[1]))
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "nodeport: Invalid end port"),
			}
			return
		}

		if start >= end {
			err = &errortypes.ParseError{
				errors.New("nodeport: Start port larger than end port"),
			}
			return
		}

		ranges = append(ranges, &PortRange{
			Start: start,
			End:   end,
		})
	}

	if len(ranges) == 0 {
		err = &errortypes.NotFoundError{
			errors.New("nodeport: No node ports configured"),
		}
		return
	}

	return
}

func New(db *database.Database, dcId primitive.ObjectID,
	resourceId primitive.ObjectID, protocol string, requestPort int) (
	ndePrt *NodePort, errData *errortypes.ErrorData, err error) {

	maxAttempts := settings.Hypervisor.NodePortMaxAttempts

	ranges, err := GetPortRanges()
	if err != nil {
		return
	}

	ndPt := &NodePort{
		Datacenter: dcId,
		Resource:   resourceId,
		Protocol:   protocol,
	}

	errData, err = ndPt.Validate(db)
	if err != nil || errData != nil {
		return
	}

	attempted := set.NewSet()
	for i := 0; i < maxAttempts; i++ {
		selectedRange := ranges[utils.RandInt(0, len(ranges)-1)]
		ndPt.Port = utils.RandInt(selectedRange.Start, selectedRange.End)

		if attempted.Contains(ndPt.Port) {
			i--
			continue
		}
		attempted.Add(ndPt.Port)

		err = ndPt.Insert(db)
		if err != nil {
			if _, ok := err.(*database.DuplicateKeyError); ok {
				err = nil
				continue
			}
			return
		}

		ndePrt = ndPt
		break
	}

	if ndePrt == nil {
		err = &errortypes.NotFoundError{
			errors.New("nodeport: No available node ports found"),
		}
		return
	}

	return
}

func RemoveResource(db *database.Database, resourceId primitive.ObjectID) (
	err error) {

	coll := db.NodePorts()

	_, err = coll.DeleteMany(db, &bson.M{
		"resource": resourceId,
	})
	if err != nil {
		err = database.ParseError(err)
		switch err.(type) {
		case *database.NotFoundError:
			err = nil
		default:
			return
		}
	}

	return
}
