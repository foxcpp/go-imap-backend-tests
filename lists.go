package backendtests

import (
	"strings"
	"testing"
)

var Blacklist []string
var Whitelist []string

func skipIfExcluded(t *testing.T) {
	if Whitelist != nil {
		whitelisted := false
		for _, included := range Whitelist {
			if strings.HasPrefix(t.Name(), included) {
				whitelisted = true
			}
		}
		if !whitelisted {
			t.Skip("not in whitelist")
			t.SkipNow()
		}
	}

	for _, excluded := range Blacklist {
		if strings.HasPrefix(t.Name(), excluded) {
			t.Skip("blacklisted")
			t.SkipNow()
		}
	}
}
