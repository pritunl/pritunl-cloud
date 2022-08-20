package main

import (
	"testing"

	"github.com/pritunl/pritunl-cloud/cmd"
	"github.com/pritunl/pritunl-cloud/constants"
)

func TestServer(t *testing.T) {
	constants.Production = false

	Init()
	err := cmd.Node(true)
	if err != nil {
		panic(err)
	}

	return
}
