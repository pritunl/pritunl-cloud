package utils

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

var (
	numRe = regexp.MustCompile(`(\d+|\D+)`)
)

type ObjectIdSlice []primitive.ObjectID

func (o ObjectIdSlice) Len() int {
	return len(o)
}

func (o ObjectIdSlice) Less(i, j int) bool {
	return o[i].Hex() < o[j].Hex()
}

func (o ObjectIdSlice) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func SortObjectIds(x []primitive.ObjectID) {
	sort.Sort(ObjectIdSlice(x))
}

func NaturalCompare(a, b string) int {
	aParts := numRe.FindAllString(a, -1)
	bParts := numRe.FindAllString(b, -1)

	minLen := len(aParts)
	if len(bParts) < minLen {
		minLen = len(bParts)
	}

	for i := 0; i < minLen; i++ {
		aPart := aParts[i]
		bPart := bParts[i]

		aNum, aErr := strconv.Atoi(aPart)
		bNum, bErr := strconv.Atoi(bPart)

		if aErr == nil && bErr == nil {
			if aNum != bNum {
				return aNum - bNum
			}
		} else if (aErr == nil) != (bErr == nil) {
			if aErr == nil {
				return -1
			}
			return 1
		} else {
			if aPart != bPart {
				return strings.Compare(aPart, bPart)
			}
		}
	}

	return len(aParts) - len(bParts)
}
