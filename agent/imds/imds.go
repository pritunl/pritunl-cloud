package imds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/agent/constants"
	"github.com/pritunl/pritunl-cloud/agent/engine"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/imds/types"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/tools/logger"
)

var (
	client = &http.Client{
		Timeout: 10 * time.Second,
	}
	curSyncHash uint32
)

type Imds struct {
	Address string         `json:"address"`
	Port    int            `json:"port"`
	Secret  string         `json:"secret"`
	engine  *engine.Engine `json:"-"`
}

func (m *Imds) NewRequest(method, pth string, data interface{}) (
	req *http.Request, err error) {

	u := &url.URL{}
	u.Scheme = "http"
	u.Host = fmt.Sprintf("%s:%d", m.Address, m.Port)
	u.Path = pth

	var body io.Reader
	if data != nil {
		reqDataBuf := &bytes.Buffer{}
		err = json.NewEncoder(reqDataBuf).Encode(data)
		if err != nil {
			err = &errortypes.ParseError{
				errors.Wrap(err, "agent: Failed to parse request data"),
			}
			return
		}

		body = reqDataBuf
	}

	req, err = http.NewRequest(method, u.String(), body)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "agent: Failed to create imds request"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl-imds")
	req.Header.Set("Auth-Token", m.Secret)
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return
}

func (m *Imds) Get(query string) (val string, err error) {
	req, err := m.NewRequest("GET", "/query"+query, nil)

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
				"agent: Imds server get error %d - %s",
				resp.StatusCode, body),
		}
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "agent: Imds failed to read body"),
		}
		return
	}

	val = string(data)

	return
}

type SyncResp struct {
	Hash uint32 `json:"hash"`
}

func (m *Imds) Sync() (err error) {
	data, err := GetState()
	if err != nil {
		return
	}

	req, err := m.NewRequest("PUT", "/sync", data)

	resp, err := client.Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "agent: Imds request failed"),
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
			errors.Newf("agent: Imds server sync error %d - %s",
				resp.StatusCode, body),
		}
		return
	}

	respData := &SyncResp{}
	err = json.NewDecoder(resp.Body).Decode(respData)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "agent: Failed to decode imds sync resp"),
		}
		return
	}

	if curSyncHash == 0 {
		curSyncHash = respData.Hash
	} else if respData.Hash != curSyncHash && m.engine != nil {
		curSyncHash = respData.Hash

		logger.WithFields(logger.Fields{
			"hash": int(respData.Hash),
		}).Info("agent: Running engine reload")

		SetStatus(types.Reloading)

		err = m.engine.Run(engine.Reload)
		if err != nil {
			logger.WithFields(logger.Fields{
				"hash":  int(respData.Hash),
				"error": err,
			}).Error("agent: Failed to run engine reload")
			err = nil
		}

		SetStatus(types.Running)
	}

	return
}

func (m *Imds) SetEngine(eng *engine.Engine) {
	m.engine = eng
}

func (m *Imds) RunSync() {
	go func() {
		for {
			err := m.Sync()
			if err != nil {
				logger.WithFields(logger.Fields{
					"error": err,
				}).Error("agent: Failed to sync")
			}

			time.Sleep(1 * time.Second)
		}
	}()
}

func (m *Imds) SyncStatus(status string) (err error) {
	SetStatus(status)

	err = m.Sync()
	if err != nil {
		logger.WithFields(logger.Fields{
			"status": status,
			"error":  err,
		}).Error("agent: Failed to sync status")
	}

	return
}

func (m *Imds) Init() (err error) {
	confData, err := utils.Read(constants.ImdsConfPath)
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(confData), m)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "agent: Failed to unmarshal imds conf"),
		}
		return
	}

	return
}
