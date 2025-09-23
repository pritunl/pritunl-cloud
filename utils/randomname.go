package utils

import (
	"bytes"
	"fmt"
	"math/rand"
)

var (
	randElm = []string{
		"copper",
		"argon",
		"xenon",
		"radon",
		"cobalt",
		"nickel",
		"carbon",
		"helium",
		"nitrogen",
		"radium",
		"lithium",
		"silicon",
	}
)

func RandName() (name string) {
	name = fmt.Sprintf("%s-%d", randElm[rand.Intn(len(randElm))],
		rand.Intn(8999)+1000)
	return
}

func RandIp() string {
	return fmt.Sprintf("26.197.%d.%d", rand.Intn(250)+4, rand.Intn(250)+4)
}

func RandIp6() (addr string) {
	addr = "2604:4080"
	randByt, _ := RandBytes(12)
	randHex := fmt.Sprintf("%x", randByt)

	buf := bytes.Buffer{}
	for i, run := range randHex {
		if i%4 == 0 && i != len(randHex)-1 {
			buf.WriteRune(':')
		}
		buf.WriteRune(run)
	}

	addr += buf.String()

	return
}

func RandPrivateIp() string {
	return fmt.Sprintf("10.232.%d.%d", rand.Intn(250)+4, rand.Intn(250)+4)
}

func RandPrivateIp6() (addr string) {
	addr = "fd97:7d1d"
	randByt, _ := RandBytes(12)
	randHex := fmt.Sprintf("%x", randByt)

	buf := bytes.Buffer{}
	for i, run := range randHex {
		if i%4 == 0 && i != len(randHex)-1 {
			buf.WriteRune(':')
		}
		buf.WriteRune(run)
	}

	addr += buf.String()

	return
}
