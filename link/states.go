package link

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	caches = map[primitive.ObjectID]map[string]*stateCache{}
)

type stateData struct {
	Version       string            `json:"version"`
	PublicAddress string            `json:"public_address"`
	LocalAddress  string            `json:"local_address"`
	Address6      string            `json:"address6"`
	Status        map[string]string `json:"status"`
	Errors        []string          `json:"errors"`
}

type stateCache struct {
	Timestamp time.Time
	State     *State
}

func getStateCache(vpcId primitive.ObjectID, uri string) (state *State) {
	vpcCache := caches[vpcId]
	if vpcCache != nil {
		cache := vpcCache[uri]
		if cache != nil && time.Since(cache.Timestamp) <
			time.Duration(settings.Ipsec.StateCacheTtl)*time.Second {

			state = cache.State
			return
		}
	}

	return
}

func getState(vpcId primitive.ObjectID, uri, localAddr, pubAddr, pubAddr6 string) (
	state *State, err error) {

	if constants.Interrupt {
		err = &errortypes.UnknownError{
			errors.Wrap(err, "state: Interrupt"),
		}
		return
	}

	state = &State{
		PublicAddr:  pubAddr,
		PublicAddr6: pubAddr6,
	}

	uriData, err := url.ParseRequestURI(uri)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "state: Failed to parse uri"),
		}
		return
	}

	LinkStatusLock.Lock()
	status := LinkStatus[vpcId]
	LinkStatusLock.Unlock()

	var linkStatus map[string]string
	if status != nil {
		linkStatus = status[uriData.User.Username()]
	}

	data := &stateData{
		Version:       Version,
		PublicAddress: pubAddr,
		LocalAddress:  localAddr,
		Address6:      pubAddr6,
		Status:        linkStatus,
	}
	dataBuf := &bytes.Buffer{}

	err = json.NewEncoder(dataBuf).Encode(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "state: Failed to parse request data"),
		}
		return
	}

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("https://%s/link/state", uriData.Host),
		dataBuf,
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "state: Request init error"),
		}
		return
	}

	req.Header.Set("Content-Type", "application/json")

	hostId := uriData.User.Username()
	hostSecret, _ := uriData.User.Password()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	nonce, err := utils.RandStr(32)
	if err != nil {
		return
	}

	authStr := strings.Join([]string{
		hostId,
		timestamp,
		nonce,
		"PUT",
		"/link/state",
	}, "&")

	hashFunc := hmac.New(sha512.New, []byte(hostSecret))
	hashFunc.Write([]byte(authStr))
	rawSignature := hashFunc.Sum(nil)
	sig := base64.StdEncoding.EncodeToString(rawSignature)

	req.Header.Set("Auth-Token", hostId)
	req.Header.Set("Auth-Timestamp", timestamp)
	req.Header.Set("Auth-Nonce", nonce)
	req.Header.Set("Auth-Signature", sig)

	var client *http.Client
	if settings.Ipsec.SkipVerify {
		client = ClientInsec
	} else {
		client = ClientSec
	}

	start := time.Now()

	res, err := client.Do(req)
	if err != nil {
		state = getStateCache(vpcId, uri)

		logrus.WithFields(logrus.Fields{
			"duration":  utils.ToFixed(time.Since(start).Seconds(), 2),
			"has_cache": state != nil,
			"error":     err,
		}).Warn("state: Request failed")

		if state == nil {
			err = &errortypes.RequestError{
				errors.Wrap(err, "state: Request put error"),
			}
		} else {
			err = nil
		}
		return
	}
	defer res.Body.Close()

	if res.StatusCode >= 500 && res.StatusCode < 600 {
		state = getStateCache(vpcId, uri)
		if state == nil {
			err = &errortypes.RequestError{
				errors.Wrapf(err, "state: Bad status %n code from server",
					res.StatusCode),
			}
		} else {
			err = nil
		}
		return
	} else if res.StatusCode != 200 {
		err = &errortypes.RequestError{
			errors.Wrapf(err, "state: Bad status %n code from server",
				res.StatusCode),
		}
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "state: Failed to read response body"),
		}
		return
	}

	decBody, err := decResp(
		hostSecret,
		res.Header.Get("Cipher-IV"),
		res.Header.Get("Cipher-Signature"),
		string(body),
	)
	if err != nil {
		err = &errortypes.RequestError{
			errors.Wrap(err, "state: Failed to decrypt response"),
		}
		return
	}

	err = json.Unmarshal(decBody, state)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "state: Failed to unmarshal data"),
		}
		return
	}

	cache := &stateCache{
		Timestamp: time.Now(),
		State:     state,
	}

	vpcCache := caches[vpcId]
	if vpcCache == nil {
		vpcCache = map[string]*stateCache{}
		caches[vpcId] = vpcCache
	}

	vpcCache[uri] = cache

	return
}

func GetStates(vpcId primitive.ObjectID, uris []string,
	localAddr, pubAddr, pubAddr6 string) (states []*State) {

	states = []*State{}
	urisSet := set.NewSet()

	for _, uri := range uris {
		urisSet.Add(uri)

		state, err := getState(vpcId, uri, localAddr, pubAddr, pubAddr6)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"uri":   uri,
				"error": err,
			}).Info("state: Failed to get state")
			continue
		}

		states = append(states, state)
	}

	vpcCache := caches[vpcId]
	if vpcCache != nil {
		for uri := range vpcCache {
			if !urisSet.Contains(uri) {
				delete(vpcCache, uri)
			}
		}
	}

	return
}
