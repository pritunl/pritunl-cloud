package main

import (
	"github.com/pritunl/pritunl-cloud/imds/server/router"
)

func main() {
	routr := &router.Router{}
	routr.Init()

	err := routr.Run()
	if err != nil {
		panic(err)
	}
}
