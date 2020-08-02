package backendtests

import (
	"testing"

	"github.com/emersion/go-imap/backend"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

func User_Username(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	err := b.CreateUser("username1")
	assert.NilError(t, err)
	u, err := b.GetUser("username1")
	assert.NilError(t, err)
	defer assert.NilError(t, u.Logout())

	assert.Equal(t, u.Username(), "username1", "Username mismatch")
}

func User_CreateMailbox(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	err := b.CreateUser("username1")
	assert.NilError(t, err)
	u, err := b.GetUser("username1")
	assert.NilError(t, err)
	defer assert.NilError(t, u.Logout())

	assert.NilError(t, u.CreateMailbox("TESTBOX"))

	mboxes, err := u.ListMailboxes(false)
	assert.NilError(t, err)
	testBoxExists := false
	for _, mbox := range mboxes {
		if mbox.Name == "TESTBOX" {
			testBoxExists = true
		}
	}
	assert.Assert(t, testBoxExists)

	_, mbox, err := u.GetMailbox("TESTBOX", true, &noopConn{})
	assert.NilError(t, err)

	assert.Equal(t, mbox.Name(), "TESTBOX", "Mailbox name mismatch")
}

func User_CreateMailbox_Parents(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	err := b.CreateUser("username1")
	assert.NilError(t, err)
	u, err := b.GetUser("username1")
	assert.NilError(t, err)
	defer assert.NilError(t, u.Logout())

	assert.NilError(t, u.CreateMailbox("INBOX.FOOBAR.BAR"))

	mboxes, err := u.ListMailboxes(false)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(mboxes, 3), "Unexpected length of mailboxes list after mailbox creation")

	_, mbox, err := u.GetMailbox("INBOX.FOOBAR.BAR", true, &noopConn{})
	assert.NilError(t, err)
	assert.Equal(t, mbox.Name(), "INBOX.FOOBAR.BAR", "Mailbox name mismatch")

	_, mbox, err = u.GetMailbox("INBOX.FOOBAR", true, &noopConn{})
	assert.NilError(t, err)
	assert.Equal(t, mbox.Name(), "INBOX.FOOBAR", "Mailbox name mismatch")

	_, mbox, err = u.GetMailbox("INBOX", true, &noopConn{})
	assert.NilError(t, err)
	assert.Equal(t, mbox.Name(), "INBOX", "Mailbox name mismatch")
}

func User_DeleteMailbox(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	err := b.CreateUser("username1")
	assert.NilError(t, err)
	u, err := b.GetUser("username1")
	assert.NilError(t, err)
	defer assert.NilError(t, u.Logout())

	assert.NilError(t, u.CreateMailbox("TEST"))
	assert.NilError(t, u.DeleteMailbox("TEST"))
	assert.Error(t, u.DeleteMailbox("TEST"), backend.ErrNoSuchMailbox.Error(), "User.DeleteMailbox succeed")
}

func User_DeleteMailbox_Parents(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	err := b.CreateUser("username1")
	assert.NilError(t, err)
	u, err := b.GetUser("username1")
	assert.NilError(t, err)
	defer assert.NilError(t, u.Logout())

	assert.NilError(t, u.CreateMailbox("TEST.FOOBAR.FOO"))
	assert.NilError(t, u.DeleteMailbox("TEST"))
	_, _, err = u.GetMailbox("TEST.FOOBAR.FOO", true, &noopConn{})
	assert.NilError(t, err)
	_, _, err = u.GetMailbox("TEST.FOOBAR", true, &noopConn{})
	assert.NilError(t, err)
}

func User_RenameMailbox(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	err := b.CreateUser("username1")
	assert.NilError(t, err)
	u, err := b.GetUser("username1")
	assert.NilError(t, err)
	defer assert.NilError(t, u.Logout())

	assert.NilError(t, u.CreateMailbox("TEST"))
	assert.NilError(t, u.RenameMailbox("TEST", "TEST2"))
	_, _, err = u.GetMailbox("TEST", true, &noopConn{})
	assert.Error(t, err, backend.ErrNoSuchMailbox.Error(), "Mailbox with old name still exists")
	_, mbox, err := u.GetMailbox("TEST2", true, &noopConn{})
	assert.NilError(t, err, "Mailbox with new name doesn't exists")
	assert.Equal(t, mbox.Name(), "TEST2", "Mailbox name dismatch in returned object")
}

func User_RenameMailbox_Childrens(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	err := b.CreateUser("username1")
	assert.NilError(t, err)
	u, err := b.GetUser("username1")
	assert.NilError(t, err)

	assert.NilError(t, u.CreateMailbox("TEST.FOOBAR.BAR"))
	assert.NilError(t, u.RenameMailbox("TEST", "TEST2"))
	_, mbox, err := u.GetMailbox("TEST2.FOOBAR.BAR", true, &noopConn{})
	assert.NilError(t, err, "Mailbox children with new name doesn't exists")
	assert.Equal(t, mbox.Name(), "TEST2.FOOBAR.BAR", "Mailbox name dismatch in returned object")
	_, mbox, err = u.GetMailbox("TEST2.FOOBAR", true, &noopConn{})
	assert.NilError(t, err, "Mailbox children with new name doesn't exists")
	assert.Equal(t, mbox.Name(), "TEST2.FOOBAR", "Mailbox name dismatch in returned object")
}

func User_RenameMailbox_INBOX(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	err := b.CreateUser("username1")
	assert.NilError(t, err)
	u, err := b.GetUser("username1")
	assert.NilError(t, err)

	u.CreateMailbox("INBOX")
	assert.NilError(t, u.RenameMailbox("INBOX", "TEST2"))
	_, _, err = u.GetMailbox("INBOX", true, &noopConn{})
	assert.NilError(t, err, "INBOX doesn't exists anymore")
}
