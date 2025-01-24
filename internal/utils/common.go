package utils

import (
	"fmt"
	"strings"

	"github.com/rs/xid"
)

func GenID() string {
	return xid.New().String()
}

func GenIDShort() string {
	return xid.New().String()[15:]
}

func SlicesContains(s, sub []string) bool {
	mapS := make(map[string]bool, len(s))
	for _, str := range s {
		mapS[str] = true
	}

	for _, substr := range sub {
		if !mapS[substr] {
			return false
		}
	}

	return true
}

func SlicesContainEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	mapS := make(map[string]int)

	for _, str := range s1 {
		mapS[str]++
	}

	for _, str := range s2 {
		if _, found := mapS[str]; !found || mapS[str] < 1 {
			return false
		}
		mapS[str]--
	}

	return true
}

func PrintMap[V fmt.Stringer](m map[string]V) string {
	ret := "{"
	for k, v := range m {
		ret += fmt.Sprintf(`"%s": %s,`, k, v)
	}
	ret = strings.TrimSuffix(ret, ",")
	ret += "}"

	return ret
}

func LabelsDeepCopy(labels map[string]string) map[string]string {
	newMap := make(map[string]string)
	for key, value := range labels {
		newMap[key] = value
	}

	return newMap
}

func MergeLabels(labels ...map[string]string) map[string]string {
	ret := make(map[string]string)
	for _, label := range labels {
		for k, v := range label {
			ret[k] = v
		}
	}
	return ret
}
