package backendtests

import (
	"strings"
	"testing"

	"github.com/emersion/go-imap/backend"
	"gotest.tools/assert"
)

func Mailbox_Children(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	skipIfExcluded(t)

	be := newBack()
	defer closeBack(be)

	u := getUser(t, be)
	defer assert.NilError(t, u.Logout())

	assert.NilError(t, u.CreateMailbox("TEST"))
	assert.NilError(t, u.CreateMailbox("TESTC.TEST.FOOBAR"))

	checkMailboxChildrens(t, "TEST", ".", u)
	checkMailboxChildrens(t, "TESTC.TEST.FOOBAR", ".", u)
}
func checkMailboxChildrens(t *testing.T, name, delimiter string, u backend.User) {
	hasChildrenAttr := false
	hasNoChildrenAttr := false
	mboxes, err := u.ListMailboxes(false)
	assert.NilError(t, err)
	hasChildren := false
	for _, mbx := range mboxes {
		if mbx.Name == name {
			for _, attr := range mbx.Attributes {
				if attr == `\HasChildren` {
					hasChildrenAttr = true
				}
				if attr == `\HasNoChildren` {
					hasNoChildrenAttr = true
				}
			}
		}

		if strings.HasPrefix(mbx.Name, name+delimiter) {
			hasChildren = true
		}
	}
	if hasChildren {
		if !hasChildrenAttr {
			t.Error("\\HasChildren attribute is not present on directory with childrens")
		}
		if hasNoChildrenAttr {
			t.Error("\\HasNoChildren attribute is present on directory with childrens")
		}
	}
	if !hasChildren {
		if hasChildrenAttr {
			t.Error("\\HasChildren attribute is present on directory without childrens")
			t.FailNow()
		}
		if !hasNoChildrenAttr {
			t.Error("\\HasNoChildren attribute is not present on directory without childrens")
		}
	}
}
