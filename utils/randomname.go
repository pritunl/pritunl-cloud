package utils

import (
	"bytes"
	"fmt"
	"math/rand"
)

var (
	randOne = []string{
		"snowy",
		"restless",
		"calm",
		"ancient",
		"summer",
		"evening",
		"guarded",
		"lively",
		"thawing",
		"autumn",
		"thriving",
		"patient",
		"winter",
		"pleasant",
		"thundering",
		"elegant",
		"narrow",
		"abundant",
	}
	randTwo = []string{
		"waterfall",
		"meadow",
		"skies",
		"waves",
		"fields",
		"stars",
		"dreams",
		"refuge",
		"forest",
		"plains",
		"waters",
		"plateau",
		"thunder",
		"volcano",
		"wilderness",
		"peaks",
		"mountains",
		"vineyards",
	}
)

func RandName() (name string) {
	name = fmt.Sprintf("%s-%s-%d", randOne[rand.Intn(18)],
		randTwo[rand.Intn(18)], rand.Intn(8999)+1000)
	return
}

func RandIpv6() (addr string) {
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
