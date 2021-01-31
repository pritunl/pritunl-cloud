package iscsi

import (
	"fmt"
	"net/url"
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
