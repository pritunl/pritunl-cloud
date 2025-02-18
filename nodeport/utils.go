package nodeport

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/bridges"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/imds/server/errortypes"
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/tools/set"
)

type PortRange struct {
	Start int
	End   int
}

func GetResource(db *database.Database, resourceId primitive.ObjectID) (
	ndePrt *NodePort, err error) {

	coll := db.NodePorts()
	ndePrt = &NodePort{}

	err = coll.FindOne(db, &bson.M{
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
	resourceId primitive.ObjectID) (ndePrt *NodePort, err error) {

	maxAttempts := settings.Hypervisor.NodePortMaxAttempts

	ranges, err := GetPortRanges()
	if err != nil {
		return
	}

	ndPt := &NodePort{
		Datacenter: dcId,
		Resource:   resourceId,
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

func create() (err error) {
	err = iproute.BridgeAdd("", settings.Hypervisor.NodePortNetworkName)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link", "set",
		"dev", settings.Hypervisor.NodePortNetworkName, "up",
	)
	if err != nil {
		return
	}

	bridges.ClearCache()

	return
}

func getAddr() (addr string, err error) {
	address, _, err := iproute.AddressGetIface(
		"", settings.Hypervisor.NodePortNetworkName)
	if err != nil {
		return
	}

	if address != nil {
		addr = address.Local + fmt.Sprintf("/%d", address.Prefix)
	}

	return
}

func setAddr(addr string) (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link", "set",
		"dev", settings.Hypervisor.NodePortNetworkName, "up",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "addr", "flush",
		"dev", settings.Hypervisor.NodePortNetworkName,
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "addr", "add", addr,
		"dev", settings.Hypervisor.NodePortNetworkName,
	)
	if err != nil {
		return
	}

	return
}

func clearAddr() (err error) {
	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "link", "set",
		"dev", settings.Hypervisor.NodePortNetworkName, "up",
	)
	if err != nil {
		return
	}

	_, err = utils.ExecCombinedOutputLogged(
		nil,
		"ip", "addr", "flush",
		"dev", settings.Hypervisor.NodePortNetworkName,
	)
	if err != nil {
		return
	}

	return
}
