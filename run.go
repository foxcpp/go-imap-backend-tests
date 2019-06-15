package backendtests

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/emersion/go-imap/backend"
)

type Backend interface {
	backend.Backend
	IMAPUsersDB
}

// NewBackFunc should create new Backend object configured for testing.
//
// It should ensure that backend object created with each call gets a clean
// empty state.
type NewBackFunc func() Backend

// CloseBackFunc should clean up Backend object after testing.
//
// Most importantly, it should ensure that all persistent data is removed
// so next test will get clean state.
type CloseBackFunc func(Backend)

type testFunc func(*testing.T, NewBackFunc, CloseBackFunc)

func getFunctionName(i interface{}) string {
	parts := strings.Split(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name(), "/")
	prefix := "go-imap-backend-tests."
	return parts[len(parts)-1][len(prefix):]
}

// RunTests runs all tests against backend created using passed callback
// functions.
func RunTests(t *testing.T, newBackend NewBackFunc, closeBackend CloseBackFunc) {
	addTest := func(f testFunc) {
		t.Run(getFunctionName(f), func(t *testing.T) {
			skipIfExcluded(t)
			f(t, newBackend, closeBackend)
		})
	}

	addTest(TestInit)
	addTest(UserDB_CreateUser)
	addTest(UserDB_Login)
	addTest(UserDB_DeleteUser)
	addTest(UserDB_SetPassword)
	addTest(User_Username)
	addTest(User_CreateMailbox)
	addTest(User_CreateMailbox_Parents)
	addTest(User_DeleteMailbox)
	addTest(User_DeleteMailbox_Parents)
	addTest(User_RenameMailbox)
	addTest(User_RenameMailbox_Childrens)
	addTest(User_RenameMailbox_INBOX)
	addTest(Mailbox_Info)
	addTest(Mailbox_Children)
	addTest(Mailbox_Status)
	addTest(Mailbox_SetSubscribed)
	addTest(Mailbox_CreateMessage)
	addTest(Mailbox_UidValidity_On_Rename)
	addTest(Mailbox_ListMessages)
	addTest(Mailbox_ListMessages_Body)
	addTest(Mailbox_ListMessages_BodyPeek)
	addTest(Mailbox_ListMessages_Meta)
	addTest(Mailbox_ListMessages_Multi)
	addTest(Mailbox_FetchEncoded)
	addTest(Mailbox_MatchEncoded)
	addTest(Mailbox_SearchMessages)
	addTest(Mailbox_SetMessageFlags)
	addTest(Mailbox_MonotonicUid)
	addTest(Mailbox_Expunge)
	addTest(Mailbox_CopyMessages)

	addTest(Mailbox_ExpungeUpdate)
	addTest(Mailbox_StatusUpdate)
	addTest(Mailbox_StatusUpdate_Copy)
	addTest(Mailbox_StatusUpdate_Move)
	addTest(Mailbox_MessageUpdate)

	// MOVE extension
	addTest(Mailbox_MoveMessages)

	// APPEND-LIMIT extension
	addTest(Backend_AppendLimit)
	addTest(User_AppendLimit)
	addTest(Mailbox_AppendLimit)
}

func TestInit(t *testing.T, newBackend NewBackFunc, closeBackend CloseBackFunc) {
	b := newBackend()
	closeBackend(b)
}
