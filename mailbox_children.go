package backendtests

import (
	"strings"
	"testing"

	imap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
	"github.com/foxcpp/go-imap-backend-tests/children"
	"gotest.tools/assert"
)

func Mailbox_Children(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	u := getUser(t, b)
	defer assert.NilError(t, u.Logout())

	assert.NilError(t, u.CreateMailbox("TEST"))
	mbox, err := u.GetMailbox("TEST")
	assert.NilError(t, err)

	assert.NilError(t, u.CreateMailbox("TESTC.TEST.FOOBAR"))
	mboxC, err := u.GetMailbox("TESTC")
	assert.NilError(t, err)

	info, err := mbox.Info()
	assert.NilError(t, err)
	assert.Equal(t, info.Name, mbox.Name(), "Mailbox name mismatch")

	t.Run("HasChildren attr", func(t *testing.T) {
		b, ok := b.(children.Backend)
		if !ok || !b.EnableChildrenExt() {
			t.Skip("CHILDREN extension is not implemeted")
			t.SkipNow()
		}

		info, err := mbox.Info()
		assert.NilError(t, err)
		checkMailboxChildrens(t, info, u, mbox)

		infoC, err := mboxC.Info()
		assert.NilError(t, err)
		checkMailboxChildrens(t, infoC, u, mboxC)
	})

}
func checkMailboxChildrens(t *testing.T, info *imap.MailboxInfo, u backend.User, mbox backend.Mailbox) {
	hasChildrenAttr := false
	hasNoChildrenAttr := false
	for _, attr := range info.Attributes {
		if attr == `\HasChildren` {
			hasChildrenAttr = true
		}
		if attr == `\HasNoChildren` {
			hasNoChildrenAttr = true
		}
	}
	mboxes, err := u.ListMailboxes(false)
	assert.NilError(t, err)
	hasChildren := false
	for _, mbx := range mboxes {
		if strings.HasPrefix(mbx.Name(), info.Name+info.Delimiter) {
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
