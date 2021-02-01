package iscsi

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Device struct {
	Host     string `bson:"host" json:"host"`
	Port     int    `bson:"port" json:"port"`
	Iqn      string `bson:"iqn" json:"iqn"`
	Lun      string `bson:"lun" json:"lun"`
	Username string `bson:"username" json:"username"`
	Password string `bson:"password" json:"password"`
	Uri      string `bson:"-" json:"uri"`
}

func (d *Device) Json() {
	uri := url.URL{
		Scheme: "iscsi",
		Host:   fmt.Sprint("%s:%d", d.Host, d.Port),
		Path:   fmt.Sprintf("%s/%s", d.Iqn, d.Lun),
	}

	if d.Username != "" {
		uri.User = url.UserPassword(d.Username, d.Password)
	}

	d.Uri = uri.String()
}

func (d *Device) Parse() (errData *errortypes.ErrorData, err error) {
	if d.Uri == "" {
		errData = &errortypes.ErrorData{
			Error:   "invalid_iscsi_uri",
			Message: "Invalid iSCSI URI",
		}
		return
	}

	uri, err := url.Parse(d.Uri)
	if err != nil {
		err = nil
		errData = &errortypes.ErrorData{
			Error:   "invalid_iscsi_uri",
			Message: "Invalid iSCSI URI",
		}
		return
	}

	if uri.Scheme != "iscsi" {
		errData = &errortypes.ErrorData{
			Error:   "invalid_iscsi_uri",
			Message: "Invalid iSCSI URI",
		}
		return
	}

	port := 0
	portStr := uri.Port()
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			err = nil
			errData = &errortypes.ErrorData{
				Error:   "invalid_iscsi_port",
				Message: "Invalid iSCSI port",
			}
			return
		}
	}

	host := uri.Hostname()
	if host == "" {
		errData = &errortypes.ErrorData{
			Error:   "invalid_iscsi_uri",
			Message: "Invalid iSCSI URI",
		}
		return
	}

	username := ""
	password := ""
	if uri.User != nil {
		username = uri.User.Username()
		password, _ = uri.User.Password()

		if username != "" || password != "" {
			if username == "" {
				errData = &errortypes.ErrorData{
					Error:   "invalid_iscsi_username",
					Message: "Missing iSCSI username",
				}
				return
			}
			if password == "" {
				errData = &errortypes.ErrorData{
					Error:   "invalid_iscsi_password",
					Message: "Missing iSCSI password",
				}
				return
			}
		}
	}

	path := strings.Split(uri.Path, "/")
	if len(path) != 3 {
		errData = &errortypes.ErrorData{
			Error:   "missing_iscsi_iqn_lun",
			Message: "Missing iSCSI IQN and LUN",
		}
		return
	}

	iqn := path[1]
	if iqn == "" {
		errData = &errortypes.ErrorData{
			Error:   "invalid_iscsi_iqn",
			Message: "Invalid iSCSI IQN",
		}
		return
	}

	lun := path[2]
	if lun == "" {
		errData = &errortypes.ErrorData{
			Error:   "invalid_iscsi_lun",
			Message: "Invalid iSCSI LUN",
		}
		return
	}

	d.Host = host
	d.Port = port
	d.Iqn = iqn
	d.Lun = lun
	d.Username = username
	d.Password = password
	d.Uri = ""

	return
}
