package imds

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/server/utils"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/journal"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/store"
	pritunlutils "github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/tools/errors"
	"github.com/sirupsen/logrus"
)

var (
	hashes     = map[primitive.ObjectID]uint32{}
	hashesLock = sync.Mutex{}
	counter    = atomic.Uint64{}
)

const (
	counterMax = 2000000000
)

func Sync(db *database.Database, namespace string,
	instId, deplyId primitive.ObjectID, conf *types.Config) (err error) {

	sockPath := paths.GetImdsSockPath(instId)

	exists, err := utils.Exists(sockPath)
	if err != nil {
		return
	}

	if !exists {
		return
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context,
				_, _ string) (net.Conn, error) {

				return net.Dial("unix", sockPath)
			},
		},
		Timeout: 6 * time.Second,
	}

	var body io.Reader

	hashesLock.Lock()
	curHash := hashes[instId]
	hashesLock.Unlock()
	var newHash uint32

	if conf != nil && curHash != conf.Hash {
		newHash = conf.Hash

		reqDataBuf := &bytes.Buffer{}
		err = json.NewEncoder(reqDataBuf).Encode(conf)
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrap(err, "agent: Failed to parse request data"),
			}
			return
		}

		body = reqDataBuf
	}

	req, err := http.NewRequest("PUT", "http://unix/sync", body)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "agent: Failed to create imds request"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl-imds")
	req.Header.Set("Auth-Token", conf.ImdsHostSecret)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, e := client.Do(req)
	if e != nil {
		err = &errortypes.RequestError{
			errors.Wrap(e, "agent: Imds request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body := ""
		data, _ := ioutil.ReadAll(resp.Body)
		if data != nil {
			body = string(data)
		}

		errData := &errortypes.ErrorData{}
		err = json.Unmarshal(data, errData)
		if err != nil || errData.Error == "" {
			errData = nil
		}

		if errData != nil && errData.Message != "" {
			body = errData.Message
		}

		err = &errortypes.RequestError{
			errors.Newf(
				"agent: Imds host sync error %d - %s",
				resp.StatusCode, body),
		}
		return
	}

	ste := &types.State{}
	err = json.NewDecoder(resp.Body).Decode(ste)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "agent: Failed to decode imds host sync resp"),
		}
		return
	}

	if newHash != 0 && ste.Status != "" {
		hashesLock.Lock()
		hashes[instId] = newHash
		hashesLock.Unlock()
	}

	if ste.Status != "" {
		coll := db.Instances()

		data := bson.M{
			"guest.status":    ste.Status,
			"guest.timestamp": time.Now(),
			"guest.heartbeat": ste.Timestamp,
			"guest.memory":    ste.Memory,
			"guest.hugepages": ste.HugePages,
			"guest.load1":     ste.Load1,
			"guest.load5":     ste.Load5,
			"guest.load15":    ste.Load15,
		}

		if ste.DhcpIp != nil {
			data["dhcp_ip"] = ste.DhcpIp.String()
		}
		if ste.DhcpIp6 != nil {
			data["dhcp_ip6"] = ste.DhcpIp6.String()
		}

		if ste.Updates != nil {
			data["guest.updates"] = ste.Updates
		}

		_, err = coll.UpdateOne(db, &bson.M{
			"_id": instId,
		}, bson.M{
			"$set": data,
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}

		var kind int32
		var resource primitive.ObjectID
		if !deplyId.IsZero() {
			kind = journal.DeploymentAgent
			resource = deplyId
		} else {
			kind = journal.InstanceAgent
			resource = instId
		}

		for _, entry := range ste.Output {
			jrnl := &journal.Journal{
				Resource:  resource,
				Kind:      kind,
				Level:     entry.Level,
				Timestamp: entry.Timestamp,
				Count:     int32(counter.Add(1) % counterMax),
				Message:   entry.Message,
			}

			err = jrnl.Insert(db)
			if err != nil {
				return
			}
		}

		curIp := ""
		curIp6 := ""
		curIpPrefix := ""
		curIpPrefix6 := ""
		curIpCached := false
		curIpCached6 := false
		newIp := ""
		newIp6 := ""
		newIpPrefix := ""
		newIpPrefix6 := ""
		clearIpCache := false

		if ste.DhcpIp != nil {
			addrStore, ok := store.GetAddress(instId)
			if ok {
				curIpCached = true
				curIp = addrStore.Addr
				if ste.DhcpIface == ste.DhcpIface6 {
					curIpCached6 = true
					curIp6 = addrStore.Addr6
				}
			} else {
				address, address6, e := iproute.AddressGetIfaceMod(
					namespace, ste.DhcpIface)
				if e == nil && address != nil {
					curIpCached = false
					curIp = address.Local
					curIpPrefix = fmt.Sprintf(
						"%s/%d", address.Local, address.Prefix)
					if ste.DhcpIface == ste.DhcpIface6 && address6 != nil {
						curIpCached6 = false
						curIp6 = address6.Local
						curIpPrefix6 = fmt.Sprintf(
							"%s/%d", address6.Local, address6.Prefix)
					}
				}
			}
			newIpPrefix = ste.DhcpIp.String()
			newIp = strings.Split(newIpPrefix, "/")[0]
		}

		if ste.DhcpIp6 != nil {
			if curIp6 == "" {
				addrStore, ok := store.GetAddress(instId)
				if ok {
					curIpCached6 = true
					curIp6 = addrStore.Addr6
				} else {
					_, address6, e := iproute.AddressGetIfaceMod(
						namespace, ste.DhcpIface6)
					if e == nil && address6 != nil {
						curIpCached6 = false
						curIp6 = address6.Local
						curIpPrefix6 = fmt.Sprintf(
							"%s/%d", address6.Local, address6.Prefix)
					}
				}
			}
			newIpPrefix6 = ste.DhcpIp6.String()
			newIp6 = strings.Split(newIpPrefix6, "/")[0]
		}

		if newIp != "" && newIp != curIp {
			if curIpCached {
				address, address6, e := iproute.AddressGetIfaceMod(
					namespace, ste.DhcpIface)
				if e == nil && address != nil {
					curIpCached = false
					curIp = address.Local
					curIpPrefix = fmt.Sprintf(
						"%s/%d", address.Local, address.Prefix)
					if ste.DhcpIface == ste.DhcpIface6 && address6 != nil {
						curIpCached6 = false
						curIp6 = address6.Local
						curIpPrefix6 = fmt.Sprintf(
							"%s/%d", address6.Local, address6.Prefix)
					}
				}
			}

			if newIp != curIp {
				logrus.WithFields(logrus.Fields{
					"instance":  instId.Hex(),
					"namespace": namespace,
					"cur_ip":    curIpPrefix,
					"new_ip":    newIpPrefix,
				}).Info("imds: Updating instance DHCP IPv4 address")

				if curIpPrefix != "" {
					_, err = pritunlutils.ExecCombinedOutputLogged(
						[]string{"File exists", "Cannot assign"},
						"ip", "netns", "exec", namespace,
						"ip", "addr",
						"del", curIpPrefix,
						"dev", ste.DhcpIface,
					)
					if err != nil {
						return
					}
				}
				_, err = pritunlutils.ExecCombinedOutputLogged(
					[]string{"File exists", "already assigned"},
					"ip", "netns", "exec", namespace,
					"ip", "addr",
					"add", newIpPrefix,
					"dev", ste.DhcpIface,
				)
				if err != nil {
					return
				}

				if ste.DhcpGateway != nil {
					_, err = pritunlutils.ExecCombinedOutputLogged(
						[]string{"File exists"},
						"ip", "netns", "exec", namespace,
						"ip", "route",
						"add", "default",
						"via", ste.DhcpGateway.String(),
						"dev", ste.DhcpIface,
					)
					if err != nil {
						return
					}
				}
				clearIpCache = true
			}
		}

		if newIp6 != "" && newIp6 != curIp6 {
			if curIpCached6 {
				_, address6, e := iproute.AddressGetIfaceMod(
					namespace, ste.DhcpIface6)
				if e == nil && address6 != nil {
					curIpCached6 = false
					curIp6 = address6.Local
					curIpPrefix6 = fmt.Sprintf(
						"%s/%d", address6.Local, address6.Prefix)
				}
			}

			if newIp6 != curIp6 {
				logrus.WithFields(logrus.Fields{
					"instance":  instId.Hex(),
					"namespace": namespace,
					"cur_ip6":   curIpPrefix6,
					"new_ip6":   newIpPrefix6,
				}).Info("imds: Updating instance DHCP IPv6 address")

				if curIpPrefix6 != "" {
					_, err = pritunlutils.ExecCombinedOutputLogged(
						[]string{"File exists", "Cannot assign"},
						"ip", "netns", "exec", namespace,
						"ip", "addr",
						"del", curIpPrefix6,
						"dev", ste.DhcpIface6,
					)
					if err != nil {
						return
					}
				}
				_, err = pritunlutils.ExecCombinedOutputLogged(
					[]string{"File exists", "already assigned"},
					"ip", "netns", "exec", namespace,
					"ip", "addr",
					"add", newIpPrefix6,
					"dev", ste.DhcpIface6,
				)
				if err != nil {
					return
				}
				clearIpCache = true
			}
		}

		if clearIpCache {
			store.RemAddress(instId)
		}
	}

	return
}

func Pull(db *database.Database, instId, deplyId primitive.ObjectID,
	imdsHostSecret string) (err error) {

	sockPath := paths.GetImdsSockPath(instId)

	exists, err := utils.Exists(sockPath)
	if err != nil {
		return
	}

	if !exists {
		return
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context,
				_, _ string) (net.Conn, error) {

				return net.Dial("unix", sockPath)
			},
		},
		Timeout: 6 * time.Second,
	}

	req, err := http.NewRequest("GET", "http://unix/sync", nil)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "agent: Failed to create imds request"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl-imds")
	req.Header.Set("Auth-Token", imdsHostSecret)

	resp, e := client.Do(req)
	if e != nil {
		err = &errortypes.RequestError{
			errors.Wrap(e, "agent: Imds request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body := ""
		data, _ := ioutil.ReadAll(resp.Body)
		if data != nil {
			body = string(data)
		}

		errData := &errortypes.ErrorData{}
		err = json.Unmarshal(data, errData)
		if err != nil || errData.Error == "" {
			errData = nil
		}

		if errData != nil && errData.Message != "" {
			body = errData.Message
		}

		err = &errortypes.RequestError{
			errors.Newf(
				"agent: Imds host sync error %d - %s",
				resp.StatusCode, body),
		}
		return
	}

	ste := &types.State{}
	err = json.NewDecoder(resp.Body).Decode(ste)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "agent: Failed to decode imds host sync resp"),
		}
		return
	}

	if ste.Status != "" {
		coll := db.Instances()

		data := bson.M{
			"guest.status":    ste.Status,
			"guest.timestamp": time.Now(),
			"guest.heartbeat": ste.Timestamp,
			"guest.memory":    ste.Memory,
			"guest.hugepages": ste.HugePages,
			"guest.load1":     ste.Load1,
			"guest.load5":     ste.Load5,
			"guest.load15":    ste.Load15,
		}

		if ste.DhcpIp != nil {
			data["dhcp_ip"] = ste.DhcpIp.String()
		}
		if ste.DhcpIp6 != nil {
			data["dhcp_ip6"] = ste.DhcpIp6.String()
		}

		if ste.Updates != nil {
			data["guest.security"] = ste.Updates
		}

		_, err = coll.UpdateOne(db, &bson.M{
			"_id": instId,
		}, bson.M{
			"$set": &bson.M{
				"guest": data,
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}

		var kind int32
		var resource primitive.ObjectID
		if !deplyId.IsZero() {
			kind = journal.DeploymentAgent
			resource = deplyId
		} else {
			kind = journal.InstanceAgent
			resource = instId
		}

		for _, entry := range ste.Output {
			jrnl := &journal.Journal{
				Resource:  resource,
				Kind:      kind,
				Level:     entry.Level,
				Timestamp: entry.Timestamp,
				Count:     int32(counter.Add(1) % counterMax),
				Message:   entry.Message,
			}

			err = jrnl.Insert(db)
			if err != nil {
				return
			}
		}
	}

	return
}

func State(db *database.Database, instId primitive.ObjectID,
	imdsHostSecret string) (ste *types.State, err error) {

	sockPath := paths.GetImdsSockPath(instId)

	exists, err := utils.Exists(sockPath)
	if err != nil {
		return
	}

	if !exists {
		return
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context,
				_, _ string) (net.Conn, error) {

				return net.Dial("unix", sockPath)
			},
		},
		Timeout: 6 * time.Second,
	}

	req, err := http.NewRequest("GET", "http://unix/state", nil)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "agent: Failed to create imds request"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl-imds")
	req.Header.Set("Auth-Token", imdsHostSecret)

	resp, e := client.Do(req)
	if e != nil {
		err = &errortypes.RequestError{
			errors.Wrap(e, "agent: Imds request failed"),
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body := ""
		data, _ := ioutil.ReadAll(resp.Body)
		if data != nil {
			body = string(data)
		}

		errData := &errortypes.ErrorData{}
		err = json.Unmarshal(data, errData)
		if err != nil || errData.Error == "" {
			errData = nil
		}

		if errData != nil && errData.Message != "" {
			body = errData.Message
		}

		err = &errortypes.RequestError{
			errors.Newf(
				"agent: Imds host sync error %d - %s",
				resp.StatusCode, body),
		}
		return
	}

	ste = &types.State{}
	err = json.NewDecoder(resp.Body).Decode(ste)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "agent: Failed to decode imds host sync resp"),
		}
		return
	}

	return
}
