package imds

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/agent/constants"
	"github.com/pritunl/pritunl-cloud/agent/logging"
	"github.com/pritunl/pritunl-cloud/engine"
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
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Secret      string            `json:"secret"`
	State       string            `json:"state"`
	engine      *engine.Engine    `json:"-"`
	initialized bool              `json:"-"`
	waiter      sync.WaitGroup    `json:"-"`
	syncLock    sync.Mutex        `json:"-"`
	logger      *logging.Redirect `json:"-"`
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
	query = strings.TrimPrefix(query, "+")

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
	Spec string `json:"spec"`
	Hash uint32 `json:"hash"`
}

func (m *Imds) SyncReady(timeout time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	var lastErr error
	for {
		select {
		case <-ctx.Done():
			if lastErr != nil {
				err = lastErr
			} else {
				err = &errortypes.RequestError{
					errors.New("agent: Initial config timeout"),
				}
			}
			return
		case <-ticker.C:
			ready, e := m.Sync()
			if e != nil {
				lastErr = e
				continue
			}
			if !ready {
				continue
			}
			return nil
		}
	}
}

func (m *Imds) Sync() (ready bool, err error) {
	m.syncLock.Lock()
	defer m.syncLock.Unlock()

	data, err := m.GetState(curSyncHash)
	if err != nil {
		return
	}

	if m.logger != nil {
		data.Output = m.logger.GetOutput()
	}

	req, err := m.NewRequest("PUT", "/sync", data)
	if err != nil {
		return
	}

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

	ready = true
	if respData.Hash == 0 {
		ready = false
	} else if curSyncHash == 0 {
		curSyncHash = respData.Hash
	} else if respData.Hash != curSyncHash && respData.Spec != "" &&
		m.engine != nil && m.initialized {

		curSyncHash = respData.Hash

		logger.WithFields(logger.Fields{
			"spec_len": len(respData.Spec),
			"hash":     int(respData.Hash),
		}).Info("agent: Running engine reload")

		SetStatus(types.Reloading)

		err = m.engine.UpdateSpec(respData.Spec)
		if err != nil {
			logger.WithFields(logger.Fields{
				"spec_len": len(respData.Spec),
				"hash":     int(respData.Hash),
				"error":    err,
			}).Error("agent: Failed to run engine spec update")
			err = nil
		}

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

func (m *Imds) SetInitialized() {
	m.initialized = true
}

func (m *Imds) RunSync(fast bool) {
	m.waiter.Add(1)

	go func() {
		defer m.waiter.Done()

		for {
			_, err := m.Sync()
			if err != nil {
				logger.WithFields(logger.Fields{
					"error": err,
				}).Error("agent: Failed to sync")
			}

			if fast {
				time.Sleep(500 * time.Millisecond)
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}()
}

func (m *Imds) SyncStatus(status string) (err error) {
	SetStatus(status)

	_, err = m.Sync()
	if err != nil {
		logger.WithFields(logger.Fields{
			"status": status,
			"error":  err,
		}).Error("agent: Failed to sync status")
	}

	return
}

func (m *Imds) Wait() (err error) {
	m.waiter.Wait()

	return
}

func (m *Imds) Init(eng *engine.Engine) (err error) {
	m.engine = eng

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

func (m *Imds) OpenLog() (err error) {
	m.logger = &logging.Redirect{}

	err = m.logger.Open()
	if err != nil {
		return
	}

	return
}

func (m *Imds) Close() {
	if m.logger != nil {
		m.logger.Close()
	}
}
