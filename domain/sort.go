package domain

import (
	"fmt"
	"strings"
)

type Records []*Record

func (r Records) Len() int {
	return len(r)
}

func (r Records) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r Records) Less(i, j int) bool {
	partsI := strings.Split(r[i].SubDomain, ".")
	partsJ := strings.Split(r[j].SubDomain, ".")

	fmt.Println(partsI)

	for idx := 0; idx < len(partsI)/2; idx++ {
		partsI[idx], partsI[len(partsI)-1-idx] = partsI[len(
			partsI)-1-idx], partsI[idx]
	}
	for idx := 0; idx < len(partsJ)/2; idx++ {
		partsJ[idx], partsJ[len(partsJ)-1-idx] = partsJ[len(
			partsJ)-1-idx], partsJ[idx]
	}

	minLen := len(partsI)
	if len(partsJ) < minLen {
		minLen = len(partsJ)
	}

	for idx := 0; idx < minLen; idx++ {
		if partsI[idx] != partsJ[idx] {
			return partsI[idx] < partsJ[idx]
		}
	}

	return len(partsI) < len(partsJ)
}
