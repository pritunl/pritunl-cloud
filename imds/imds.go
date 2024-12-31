package imds

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/server/utils"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/journal"
	"github.com/pritunl/pritunl-cloud/paths"
	"github.com/pritunl/tools/errors"
)

var (
	hashes     = map[primitive.ObjectID]uint32{}
	hashesLock = sync.Mutex{}
)

func Sync(db *database.Database, instId, deplyId primitive.ObjectID,
	conf *types.Config) (err error) {

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
		hashes[instId] = curHash
		hashesLock.Unlock()
	}

	if ste.Status != "" {
		coll := db.Instances()

		_, err = coll.UpdateOne(db, &bson.M{
			"_id": instId,
		}, bson.M{
			"$set": &bson.M{
				"guest": &instance.GuestData{
					Status:    ste.Status,
					Heartbeat: ste.Timestamp,
					Memory:    ste.Memory,
					HugePages: ste.HugePages,
					Load1:     ste.Load1,
					Load5:     ste.Load5,
					Load15:    ste.Load15,
				},
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}

		var kind int
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
				Timestamp: entry.Timestamp,
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

		_, err = coll.UpdateOne(db, &bson.M{
			"_id": instId,
		}, bson.M{
			"$set": &bson.M{
				"guest": &instance.GuestData{
					Status:    ste.Status,
					Heartbeat: ste.Timestamp,
					Memory:    ste.Memory,
					HugePages: ste.HugePages,
					Load1:     ste.Load1,
					Load5:     ste.Load5,
					Load15:    ste.Load15,
				},
			},
		})
		if err != nil {
			err = database.ParseError(err)
			return
		}

		var kind int
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
				Timestamp: entry.Timestamp,
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
