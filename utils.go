package backendtests

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
	"github.com/google/go-cmp/cmp"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

func getNamedUser(t *testing.T, b Backend, name string) backend.User {
	t.Helper()
	err := b.CreateUser(name)
	assert.NilError(t, err)
	u, err := b.GetUser(name)
	assert.NilError(t, err)
	return u
}

func getUser(t *testing.T, b Backend) backend.User {
	t.Helper()
	name := fmt.Sprintf("test%v", time.Now().UnixNano())
	return getNamedUser(t, b, name)
}

type noopConn struct{}

func (*noopConn) SendUpdate(backend.Update) error { return nil }

func getNamedMbox(t *testing.T, u backend.User, name string, conn backend.Conn) backend.Mailbox {
	t.Helper()
	assert.NilError(t, u.CreateMailbox(name))
	if conn == nil {
		conn = &noopConn{}
	}
	_, mbox, err := u.GetMailbox(name, false, conn)
	assert.NilError(t, err)
	return mbox
}

func getMbox(t *testing.T, u backend.User, conn backend.Conn) backend.Mailbox {
	t.Helper()
	name := fmt.Sprintf("test%v", time.Now().UnixNano())
	return getNamedMbox(t, u, name, conn)
}

var baseDate = time.Time{}

func createMsgs(t *testing.T, mbox backend.Mailbox, user backend.User, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		assert.NilError(t, user.CreateMessage(mbox.Name(),
			[]string{
				"$Test" + strconv.Itoa(i+1) + "-1",
				"$Test" + strconv.Itoa(i+1) + "-2",
			},
			baseDate.Add(time.Duration((i+1)*24)*time.Hour),
			strings.NewReader(testMailString),
			mbox,
		))
		assert.NilError(t, mbox.Poll(true))
	}
}

func createMsgsUids(t *testing.T, mbox backend.Mailbox, user backend.User, count int) (res []uint32) {
	t.Helper()
	for i := 0; i < count; i++ {
		stat, err := user.Status(mbox.Name(), []imap.StatusItem{imap.StatusUidNext})
		assert.NilError(t, err)
		res = append(res, stat.UidNext)

		assert.NilError(t, user.CreateMessage(mbox.Name(),
			[]string{
				"$Test" + strconv.Itoa(i+1) + "-1",
				"$Test" + strconv.Itoa(i+1) + "-2",
			},
			baseDate.Add(time.Duration((i+1)*24)*time.Hour),
			strings.NewReader(testMailString),
			mbox,
		))
		assert.NilError(t, mbox.Poll(true))
	}
	return
}

func isNthMsg(msg *imap.Message, indx int, args ...cmp.Option) is.Comparison {
	indx = indx - 1

	msgDate := msg.InternalDate.Truncate(time.Second)
	nthDate := baseDate.Add(time.Duration(indx+1) * 24 * time.Hour).Truncate(time.Second)

	return is.DeepEqual(msgDate, nthDate, args...)
}

func isNthMsgFlags(msg *imap.Message, indx int, args ...cmp.Option) is.Comparison {
	sort.Strings(msg.Flags)
	flags := []string{
		"$Test" + strconv.Itoa(indx) + "-1",
		"$Test" + strconv.Itoa(indx) + "-2",
		imap.RecentFlag,
	}
	return is.DeepEqual(msg.Flags, []string{flags[0], flags[1], imap.RecentFlag})
}

func init() {
	if os.Getenv("SHUFFLE_CASES") == "1" {
		rand.Seed(time.Now().Unix())
	}
}
