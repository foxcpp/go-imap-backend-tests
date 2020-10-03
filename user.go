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

	mboxes, err := u.ListMailboxes(false)
	assert.NilError(t, err)

	delimiter := "."
	initialLength := 0
	if len(mboxes) > 0 {
		delimiter = mboxes[0].Delimiter
		initialLength = len(mboxes)
	}

	mailboxName := "INBOX" + delimiter + "FOOBAR" + delimiter + "BAR"
	assert.NilError(t, u.CreateMailbox(mailboxName))

	mboxes, err = u.ListMailboxes(false)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(mboxes, initialLength+2), "Unexpected length of mailboxes list after mailbox creation")

	_, mbox, err := u.GetMailbox(mailboxName, true, &noopConn{})
	assert.NilError(t, err)
	assert.Equal(t, mbox.Name(), mailboxName, "Mailbox name mismatch")

	mailboxName = "INBOX" + delimiter + "FOOBAR"
	_, mbox, err = u.GetMailbox(mailboxName, true, &noopConn{})
	assert.NilError(t, err)
	assert.Equal(t, mbox.Name(), mailboxName, "Mailbox name mismatch")

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

	mboxes, err := u.ListMailboxes(false)
	assert.NilError(t, err)

	delimiter := "."
	if len(mboxes) > 0 {
		delimiter = mboxes[0].Delimiter
	}

	mailboxName := "TEST" + delimiter + "FOOBAR" + delimiter + "FOO"
	assert.NilError(t, u.CreateMailbox(mailboxName))
	assert.NilError(t, u.DeleteMailbox("TEST"))

	_, _, err = u.GetMailbox(mailboxName, true, &noopConn{})
	assert.NilError(t, err)

	_, _, err = u.GetMailbox("TEST"+delimiter+"FOOBAR", true, &noopConn{})
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

	mboxes, err := u.ListMailboxes(false)
	assert.NilError(t, err)

	delimiter := "."
	if len(mboxes) > 0 {
		delimiter = mboxes[0].Delimiter
	}

	assert.NilError(t, u.CreateMailbox("TEST"+delimiter+"FOOBAR"+delimiter+"BAR"))
	assert.NilError(t, u.RenameMailbox("TEST", "TEST2"))

	mailboxName := "TEST2" + delimiter + "FOOBAR" + delimiter + "BAR"
	_, mbox, err := u.GetMailbox(mailboxName, true, &noopConn{})
	assert.NilError(t, err, "Mailbox children with new name doesn't exists")
	assert.Equal(t, mbox.Name(), mailboxName, "Mailbox name mismatch in returned object")

	mailboxName = "TEST2" + delimiter + "FOOBAR"
	_, mbox, err = u.GetMailbox(mailboxName, true, &noopConn{})
	assert.NilError(t, err, "Mailbox children with new name doesn't exists")
	assert.Equal(t, mbox.Name(), mailboxName, "Mailbox name mismatch in returned object")
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
