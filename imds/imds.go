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

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/dhcpc"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/imds/server/utils"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/iproute"
	"github.com/pritunl/pritunl-cloud/journal"
	"github.com/pritunl/pritunl-cloud/manifest"
	"github.com/pritunl/pritunl-cloud/metric"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/store"
	"github.com/pritunl/pritunl-cloud/systemd"
	"github.com/pritunl/pritunl-cloud/telemetry"
	"github.com/pritunl/pritunl-cloud/unit"
	pritunlutils "github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/tools/errors"
	"github.com/sirupsen/logrus"
)

var (
	hashes              = map[bson.ObjectID]uint32{}
	hashesLock          = sync.Mutex{}
	dhcpFails           = map[bson.ObjectID]time.Time{}
	dhcpFailsLock       = sync.Mutex{}
	versionRestarts     = map[bson.ObjectID]time.Time{}
	versionRestartsLock = sync.Mutex{}
	counter             = atomic.Uint64{}
)

const (
	counterMax        = 2000000000
	versionRestartTtl = 1 * time.Hour
)

func Sync(db *database.Database, namespace string,
	instId, orgId, deplyId bson.ObjectID, conf *types.Config) (err error) {

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

	if conf != nil && curHash != conf.Hash {
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

	hashesLock.Lock()
	hashes[instId] = ste.Hash
	hashesLock.Unlock()

	if ste.Version != constants.ImdsVersion {
		versionRestartsLock.Lock()
		lastRestart := versionRestarts[instId]
		restart := lastRestart.IsZero() ||
			time.Since(lastRestart) > versionRestartTtl
		if restart {
			versionRestarts[instId] = time.Now()
		}
		versionRestartsLock.Unlock()

		if restart {
			logrus.WithFields(logrus.Fields{
				"instance":     instId.Hex(),
				"imds_version": ste.Version,
				"node_version": constants.ImdsVersion,
			}).Warn("imds: Imds version mismatch, restarting")

			e := Restart(instId)
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"instance": instId.Hex(),
					"error":    e,
				}).Error("imds: Failed to restart imds")
			}

			return
		}
	} else {
		versionRestartsLock.Lock()
		delete(versionRestarts, instId)
		versionRestartsLock.Unlock()
	}

	if ste.Status != "" {
		data := bson.M{
			"status":    ste.Status,
			"timestamp": time.Now(),
			"heartbeat": ste.Timestamp,
		}

		instFields := bson.M{}
		if ste.DhcpIp != nil {
			dhcpIp := ste.DhcpIp.String()
			if conf.Instance == nil || conf.Instance.DhcpIp != dhcpIp {
				instFields["dhcp_ip"] = dhcpIp
			}
		}
		if ste.DhcpIp6 != nil {
			dhcpIp6 := ste.DhcpIp6.String()
			if conf.Instance == nil || conf.Instance.DhcpIp6 != dhcpIp6 {
				instFields["dhcp_ip6"] = dhcpIp6
			}
		}
		if len(ste.Metrics) > 0 {
			latest := ste.Metrics[len(ste.Metrics)-1]

			staticData := latest.StaticData()
			for key, val := range staticData {
				data[key] = val
			}

			e := metric.InsertSamples(db, instId, ste.Metrics)
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"instance": instId.Hex(),
					"error":    e,
				}).Error("imds: Failed to insert metrics")
			}
		}

		var components *telemetry.ComponentData
		if ste.Components != nil {
			components = validateComponents(db, instId, ste.Components)
			data["processes"] = components.Processes
			data["modules"] = components.Modules
			data["ports"] = components.Ports
		}

		err = instance.UpdateRuntime(db, instId, data)
		if err != nil {
			return
		}

		if len(instFields) > 0 {
			coll := db.Instances()

			_, err = coll.UpdateOne(db, &bson.M{
				"_id": instId,
			}, bson.M{
				"$set": instFields,
			})
			if err != nil {
				err = database.ParseError(err)
				return
			}
		}

		if ste.Status == types.Imaged && !deplyId.IsZero() {
			updated, e := deployment.SetImageReady(db, deplyId)
			if e != nil {
				err = e
				return
			}

			if updated {
				logrus.WithFields(logrus.Fields{
					"instance":   instId.Hex(),
					"deployment": deplyId.Hex(),
				}).Info("imds: Deployment image ready")

				event.PublishDispatch(db, "pod.change")
			}
		}

		if ste.Updates != nil {
			updts := validateUpdates(db, instId, ste.Updates)

			err = manifest.UpsertInstanceUpdates(
				db, instId, orgId, updts, components)
			if err != nil {
				return
			}
		}

		var kind int32
		var resource bson.ObjectID
		if !deplyId.IsZero() {
			kind = journal.DeploymentAgent
			resource = deplyId
		} else {
			kind = journal.InstanceAgent
			resource = instId
		}

		for _, entry := range ste.Output {
			if entry.Level < 1 || entry.Level > 9 {
				continue
			}
			if len(entry.Message) > 100000 {
				entry.Message = entry.Message[:100000]
			}

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

		if ste.Journals != nil {
			indexes := map[string]int32{}
			for _, jrnl := range conf.Journals {
				indexes[jrnl.Key] = jrnl.Index
			}

			for key, output := range ste.Journals {
				index := indexes[key]
				if index == 0 {
					continue
				}

				for _, entry := range output {
					if entry.Level < 1 || entry.Level > 9 {
						continue
					}
					if len(entry.Message) > 100000 {
						entry.Message = entry.Message[:100000]
					}

					jrnl := &journal.Journal{
						Resource:  resource,
						Kind:      index,
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
		}

		if ste.Primary && !deplyId.IsZero() && conf.Unit != nil &&
			ste.PrimaryTimestamp.After(conf.Unit.PrimaryTimestamp) {

			deply, e := deployment.Get(db, deplyId)
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"instance":   instId.Hex(),
					"deployment": deplyId.Hex(),
					"error":      e,
				}).Error("imds: Failed to get deployment for set primary")
			} else if deply.Unit.IsZero() ||
				(!conf.Unit.Id.IsZero() && deply.Unit != conf.Unit.Id) ||
				deply.State != deployment.Deployed ||
				deply.Action != "" {

				logrus.WithFields(logrus.Fields{
					"instance":   instId.Hex(),
					"deployment": deplyId.Hex(),
					"unit":       deply.Unit.Hex(),
					"state":      deply.State,
					"action":     deply.Action,
				}).Warn("imds: Deployment invalid for set primary")
			} else {
				updated, e := unit.SetPrimary(db, deply.Unit,
					deplyId, ste.PrimaryTimestamp)
				if e != nil {
					err = e
					return
				}

				if updated {
					logrus.WithFields(logrus.Fields{
						"deployment": deplyId.Hex(),
						"unit":       deply.Unit.Hex(),
					}).Info("imds: Set unit primary deployment")

					event.PublishDispatch(db, "pod.change")
				}
			}
		}

		curIp := ""
		curIp6 := ""
		curIpPrefix := ""
		curIpPrefix6 := ""
		curIpCached := false
		curIpCached6 := false
		curIpDynamic6 := false
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
						curIpDynamic6 = address6.Dynamic
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
						curIpDynamic6 = address6.Dynamic
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
						curIpDynamic6 = address6.Dynamic
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
					curIpDynamic6 = address6.Dynamic
				}
			}

			if newIp6 != curIp6 {
				logrus.WithFields(logrus.Fields{
					"instance":  instId.Hex(),
					"namespace": namespace,
					"cur_ip6":   curIpPrefix6,
					"new_ip6":   newIpPrefix6,
				}).Info("imds: Updating instance DHCP IPv6 address")

				if curIpPrefix6 != "" && !curIpDynamic6 {
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

		restartTtl := time.Duration(
			settings.Hypervisor.DhcpSyncTimeout) * time.Second

		if ste.DhcpTimestamp.IsZero() ||
			time.Since(ste.DhcpTimestamp) > restartTtl {

			dhcpcExists, _ := utils.Exists(paths.GetUnitPathDhcpc(instId))
			if dhcpcExists {
				dhcpFailsLock.Lock()
				failTimestamp, ok := dhcpFails[instId]
				if !ok {
					if ste.DhcpTimestamp.IsZero() {
						failTimestamp = time.Now()
					} else {
						failTimestamp = ste.DhcpTimestamp
					}
					dhcpFails[instId] = failTimestamp
				}
				dhcpFailsLock.Unlock()

				if time.Since(failTimestamp) > restartTtl {
					dhcpFailsLock.Lock()
					dhcpFails[instId] = time.Now()
					dhcpFailsLock.Unlock()

					unit := paths.GetUnitNameDhcpc(instId)
					unitState, _, e := systemd.GetState(unit)
					if e != nil {
						logrus.WithFields(logrus.Fields{
							"instance": instId.Hex(),
							"unit":     unit,
							"error":    e,
						}).Error("imds: Failed to get dhcpc state")
					} else if unitState == "active" {
						logrus.WithFields(logrus.Fields{
							"action":    conf.Instance.Action,
							"instance":  instId.Hex(),
							"timestamp": ste.DhcpTimestamp,
						}).Error("imds: Dhcp client sync timeout, restarting")

						e = dhcpc.Restart(instId)
						if e != nil {
							logrus.WithFields(logrus.Fields{
								"instance": instId.Hex(),
								"error":    e,
							}).Error("imds: Failed to restart dhcpc")
						}
					}
				}
			}
		} else {
			dhcpFailsLock.Lock()
			delete(dhcpFails, instId)
			dhcpFailsLock.Unlock()
		}
	}

	return
}

func Pull(db *database.Database, instId, orgId, deplyId bson.ObjectID,
	imdsHostSecret string, journals []*types.Journal) (err error) {

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
		}

		if ste.DhcpIp != nil {
			data["dhcp_ip"] = ste.DhcpIp.String()
		}
		if ste.DhcpIp6 != nil {
			data["dhcp_ip6"] = ste.DhcpIp6.String()
		}
		if len(ste.Metrics) > 0 {
			latest := ste.Metrics[len(ste.Metrics)-1]

			staticData := latest.StaticData()
			for key, val := range staticData {
				data["guest."+key] = val
			}

			e := metric.InsertSamples(db, instId, ste.Metrics)
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"instance": instId.Hex(),
					"error":    e,
				}).Error("imds: Failed to insert metrics")
			}
		}

		var components *telemetry.ComponentData
		if ste.Components != nil {
			components = validateComponents(db, instId, ste.Components)
			data["guest.processes"] = components.Processes
			data["guest.modules"] = components.Modules
			data["guest.ports"] = components.Ports
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

		if ste.Updates != nil {
			updts := validateUpdates(db, instId, ste.Updates)

			err = manifest.UpsertInstanceUpdates(
				db, instId, orgId, updts, components)
			if err != nil {
				return
			}
		}

		var kind int32
		var resource bson.ObjectID
		if !deplyId.IsZero() {
			kind = journal.DeploymentAgent
			resource = deplyId
		} else {
			kind = journal.InstanceAgent
			resource = instId
		}

		for _, entry := range ste.Output {
			if entry.Level < 1 || entry.Level > 9 {
				continue
			}
			if len(entry.Message) > 100000 {
				entry.Message = entry.Message[:100000]
			}

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

		if ste.Journals != nil {
			indexes := map[string]int32{}
			for _, jrnl := range journals {
				indexes[jrnl.Key] = jrnl.Index
			}

			for key, output := range ste.Journals {
				index := indexes[key]
				if index == 0 {
					continue
				}

				for _, entry := range output {
					if entry.Level < 1 || entry.Level > 9 {
						continue
					}
					if len(entry.Message) > 100000 {
						entry.Message = entry.Message[:100000]
					}

					jrnl := &journal.Journal{
						Resource:  resource,
						Kind:      index,
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
		}
	}

	return
}

func State(db *database.Database, instId bson.ObjectID,
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

func validateUpdates(db *database.Database, instId bson.ObjectID,
	updates []*telemetry.Update) (updts []*telemetry.Update) {

	updts = []*telemetry.Update{}
	invalid := 0
	var lastErr error

	for _, updt := range updates {
		if updt == nil {
			invalid += 1
			continue
		}

		errData, err := updt.Validate(db)
		if err != nil {
			invalid += 1
			lastErr = err
			continue
		}
		if errData != nil {
			invalid += 1
			lastErr = errData.GetError()
			continue
		}

		updts = append(updts, updt)
	}

	if invalid > 0 {
		logrus.WithFields(logrus.Fields{
			"instance": instId.Hex(),
			"count":    invalid,
			"error":    lastErr,
		}).Warn("imds: Skipped invalid updates")
	}

	return
}

func validateComponents(db *database.Database, instId bson.ObjectID,
	components *telemetry.ComponentData) (comps *telemetry.ComponentData) {

	comps, invalid := components.Validate()

	if invalid > 0 {
		logrus.WithFields(logrus.Fields{
			"instance": instId.Hex(),
			"count":    invalid,
		}).Warn("imds: Skipped invalid components")
	}

	return
}
