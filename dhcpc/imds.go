package dhcpc

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
	"github.com/pritunl/pritunl-cloud/errortypes"
)

var (
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)

type SyncData struct {
	Address  string `json:"address"`
	Gateway  string `json:"gateway"`
	Address6 string `json:"address6"`
	Gateway6 string `json:"gateway6"`
}

type Imds struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Secret  string `json:"secret"`
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
				errors.Wrap(err, "dhcpc: Failed to parse request data"),
			}
			return
		}

		body = reqDataBuf
	}

	req, err = http.NewRequest(method, u.String(), body)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "dhcpc: Failed to create imds request"),
		}
		return
	}

	req.Header.Set("User-Agent", "pritunl-dhcp")
	req.Header.Set("Auth-Token", m.Secret)
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return
}

func (m *Imds) Sync(lease *Lease) (err error) {
	req, err := m.NewRequest("PUT", "/dhcp", lease)
	if err != nil {
		return
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "dhcpc: Imds request failed"),
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
			errors.Newf("dhcpc: Imds server sync error %d - %s",
				resp.StatusCode, body),
		}
		return
	}

	return
}
