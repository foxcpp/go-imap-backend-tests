package backendtests

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-imap"
	move "github.com/emersion/go-imap-move"
	"github.com/emersion/go-imap/backend"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

type collectorConn struct {
	upds []backend.Update
}

func (c *collectorConn) SendUpdate(upd backend.Update) error {
	c.upds = append(c.upds, upd)
	return nil
}

func (c *collectorConn) discard(t *testing.T, n int) {
	t.Helper()
	if len(c.upds) < n {
		t.Fatal("Wanted to discard", n, "updates but only", len(c.upds), "were received")
	}
	c.upds = c.upds[:n+1]
}

func makeMsgSlots(count int) (res []uint32) {
	res = make([]uint32, count)
	for i := range res {
		res[i] = uint32(i + 1)
	}
	return
}

func checkExpungeEvents(t *testing.T, upds []backend.Update, slots *[]uint32, shouldBeLeft uint32) {
	t.Helper()
	if uint32(len(*slots)) == shouldBeLeft {
		return
	}
	for _, upd := range upds {
		switch upd := upd.(type) {
		case *backend.ExpungeUpdate:
			if upd.SeqNum > uint32(len(*slots)) {
				t.Errorf("Update's SeqNum is out of range: %v > %v", upd.SeqNum, len(*slots))
			} else if upd.SeqNum == 0 {
				t.Error("Update's SeqNum is zero.")
			} else {
				*slots = append((*slots)[:upd.SeqNum-1], (*slots)[upd.SeqNum:]...)
				t.Logf("Got ExpungeUpdate, SeqNum = %d, remaining slots = %d\n", upd.SeqNum, len(*slots))
				if uint32(len(*slots)) == shouldBeLeft {
					return
				}
			}
		default:
			t.Errorf("Expunge should not generate non-expunge updates (%T): %#v", upd, upd)
		}
	}
}

func Mailbox_StatusUpdate(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)

	u := getUser(t, b)
	defer assert.NilError(t, u.Logout())

	conn := collectorConn{}
	mbox := getMbox(t, u, &conn)
	defer mbox.Close()

	for i := uint32(1); i <= uint32(5); i++ {
		createMsgs(t, mbox, u,1)
		if i > uint32(len(conn.upds)) {
			t.Fatal("Missing update #", i)
		} 
		switch upd := conn.upds[i-1].(type) {
		case *backend.MailboxUpdate:
			assert.Check(t, is.Equal(upd.Messages, i), "Wrong amount of messages in mailbox reported in update")

			if _, ok := upd.Items[imap.StatusRecent]; ok {
				assert.Check(t, is.Equal(upd.Recent, i), "Wrong amount of recent messages in mailbox reported in update")
			}
		default:
			t.Errorf("Non-mailbox update sent by backend: %#v\n", upd)
		}
	}
}

func Mailbox_StatusUpdate_Copy(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)

	u := getUser(t, b)
	defer assert.NilError(t, u.Logout())

	srcMbox := getMbox(t, u, nil)
	defer srcMbox.Close()
	
	conn := collectorConn{}
	tgtMbox := getMbox(t, u, &conn)
	defer tgtMbox.Close()

	createMsgs(t, srcMbox, u,3)

	seq, _ := imap.ParseSeqSet("2:3")
	assert.NilError(t, srcMbox.CopyMessages(false, seq, tgtMbox.Name()))
	assert.NilError(t, tgtMbox.Poll(true))

	assert.Assert(t, is.Len(conn.upds, 1))
	switch upd := conn.upds[0].(type) {
	case *backend.MailboxUpdate:
		assert.Check(t, is.Equal(upd.Messages, uint32(2)), "Wrong amount of messages in mailbox reported in update")
		if _, ok := upd.Items[imap.StatusRecent]; ok {
			assert.Check(t, is.Equal(upd.Recent, uint32(2)), "Wrong amount of recent messages in mailbox reported in update")
		}
	default:
		t.Errorf("Non-mailbox update sent by backend: %#v\n", upd)
	}
}

func Mailbox_StatusUpdate_Move(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	
	u := getUser(t, b)
	defer assert.NilError(t, u.Logout())

	srcConn := collectorConn{}
	srcMbox := getMbox(t, u, &srcConn)
	defer srcMbox.Close()

	tgtConn := collectorConn{}
	tgtMbox := getMbox(t, u, &tgtConn)
	defer tgtMbox.Close()

	createMsgs(t, srcMbox, u, 3)
	srcConn.upds = nil

	moveMbox, ok := srcMbox.(move.Mailbox)
	if !ok {
		t.Skip("Backend doesn't supports MOVE (need move.Mailbox interface)")
		t.SkipNow()
	}

	seq, _ := imap.ParseSeqSet("2:3")
	assert.NilError(t, moveMbox.MoveMessages(false, seq, tgtMbox.Name()))
	assert.NilError(t, tgtMbox.Poll(false))

	// We expect 1 status update for target mailbox and two expunge updates
	// for source mailbox.
	
	assert.Assert(t, is.Len(tgtConn.upds, 1))
	mboxUpd, ok := tgtConn.upds[0].(*backend.MailboxUpdate)
	if !ok {
		t.Fatal("Non-MailboxUpdate received for target mailbox")
	}
	assert.Check(t, is.Equal(mboxUpd.Messages, uint32(2)), "Wrong amount of messages in mailbox reported in update for target")

	msgs := makeMsgSlots(3)
	for _, upd := range srcConn.upds {
		expungeUpd, ok := upd.(*backend.ExpungeUpdate)
		if !ok {
			t.Fatalf("Non-ExpungeUpdate received for source mailbox: %#v", upd)
		}

		if expungeUpd.SeqNum > uint32(len(msgs)) {
			t.Errorf("Update's SeqNum is out of range: %v > %v", expungeUpd.SeqNum, len(msgs))
		} else if expungeUpd.SeqNum == 0 {
			t.Error("Update's SeqNum is zero.")
		} else {
			t.Logf("Got ExpungeUpdate, SeqNum = %d, remaining slots = %d\n", expungeUpd.SeqNum, len(msgs))
			msgs = append(msgs[:expungeUpd.SeqNum-1], msgs[expungeUpd.SeqNum:]...)
		}
	}
	assert.Check(t, is.DeepEqual(msgs, []uint32{1}), "Wrong sequence of expunge updates received")
}

func Mailbox_MessageUpdate(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)

	u := getUser(t, b)
	defer assert.NilError(t, u.Logout())

	testFlagsUpdate := func(
		seqset string, expectedUpdates int,
		initialFlags map[uint32][]string, op imap.FlagsOp,
		opArg []string, expectedNewFlags map[uint32][]string) {

		t.Run(fmt.Sprintf("seqset=%v op=%v opArg=%v", seqset, op, opArg), func(t *testing.T) {
			skipIfExcluded(t)

			conn := collectorConn{}
			mbox := getMbox(t, u, &conn)
			defer mbox.Close()

			for i := 1; i <= len(initialFlags); i++ {
				assert.NilError(t, u.CreateMessage(mbox.Name(), initialFlags[uint32(i)], time.Now(), strings.NewReader(testMsg)))
				assert.NilError(t, mbox.Poll(true))
			}
			
			conn.upds = nil

			seq, _ := imap.ParseSeqSet(seqset)
			assert.NilError(t, mbox.UpdateMessagesFlags(false, seq, op, false, opArg))

			for i := 0; i < expectedUpdates; i++ {
				assert.Assert(t, i < len(conn.upds), "Not enough updates sent by backend")
				
				upd := conn.upds[i]
				switch upd := upd.(type) {
				case *backend.MessageUpdate:
					flags, ok := expectedNewFlags[upd.SeqNum]
					if !ok {
						t.Error("Unexpected update for SeqNum =", upd.SeqNum)
					}

					sort.Strings(flags)
					sort.Strings(upd.Flags)

					if !assert.Check(t, is.DeepEqual(flags, upd.Flags), "Flags mismatch on message %d", upd.SeqNum) {
						t.Log("upd.Flags:", upd.Flags)
						t.Log("Reference flag set:", flags)
					}
				default:
					t.Errorf("Non-message update sent by backend: %#v\n", upd)
				}
			}
		})
	}

	cases := []struct {
		seqset           string
		expectedUpdates  int
		initialFlags     map[uint32][]string
		op               imap.FlagsOp
		opArg            []string
		expectedNewFlags map[uint32][]string
	}{
		{
			"1,3,5", 3, map[uint32][]string{
				1: {"t1-1", "t1-2"},
				2: {"t2-3", "t2-4"},
				3: {"t3-5", "t3-6"},
				4: {"t4-7", "t4-8"},
				5: {"t5-9", "t5-10"},
			},
			imap.SetFlags, []string{"t0-1", "t0-2"},
			map[uint32][]string{
				1: {imap.RecentFlag, "t0-1", "t0-2"},
				3: {imap.RecentFlag, "t0-1", "t0-2"},
				5: {imap.RecentFlag, "t0-1", "t0-2"},
			},
		},
		{
			"1,3,5", 3, map[uint32][]string{
				1: {"t1-1", "t1-2"},
				2: {"t2-3", "t2-4"},
				3: {"t3-5", "t3-6"},
				4: {"t4-7", "t4-8"},
				5: {"t5-9", "t5-10"},
			},
			imap.AddFlags, []string{"t0-1", "t0-2"},
			map[uint32][]string{
				1: {imap.RecentFlag, "t0-1", "t0-2", "t1-1", "t1-2"},
				3: {imap.RecentFlag, "t0-1", "t0-2", "t3-5", "t3-6"},
				5: {imap.RecentFlag, "t0-1", "t0-2", "t5-10", "t5-9"},
			},
		},
		{
			"2,3,5", 3, map[uint32][]string{
				1: {"t1-1", "t1-2"},
				2: {"t0-0", "t2-4"},
				3: {"t3-5", "t0-0"},
				4: {"t4-7", "t4-8"},
				5: {"t0-0", "t5-10"},
			},
			imap.RemoveFlags, []string{"t0-0"},
			map[uint32][]string{
				2: {imap.RecentFlag, "t2-4"},
				3: {imap.RecentFlag, "t3-5"},
				5: {imap.RecentFlag, "t5-10"},
			},
		},
	}

	if os.Getenv("SHUFFLE_CASES") == "1" {
		rand.Shuffle(len(cases), func(i, j int) {
			cases[i], cases[j] = cases[j], cases[i]
		})
	}

	for _, case_ := range cases {
		testFlagsUpdate(case_.seqset, case_.expectedUpdates, case_.initialFlags, case_.op, case_.opArg, case_.expectedNewFlags)
	}
}

func Mailbox_ExpungeUpdate(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)

	u := getUser(t, b)
	defer assert.NilError(t, u.Logout())

	testSlots := func(msgsCount int, seqset string, matchedMsgs int, expectedSlots []uint32) {
		t.Run(seqset, func(t *testing.T) {
			skipIfExcluded(t)
			
			conn := collectorConn{}

			mbox := getMbox(t, u, &conn)
			defer mbox.Close()
			createMsgs(t, mbox, u, msgsCount)
			msgs := makeMsgSlots(msgsCount)

			conn.upds = nil

			seq, _ := imap.ParseSeqSet(seqset)
			assert.NilError(t, mbox.UpdateMessagesFlags(false, seq, imap.AddFlags, true, []string{imap.DeletedFlag}))

			assert.NilError(t, mbox.Expunge())
			checkExpungeEvents(t, conn.upds, &msgs, uint32(msgsCount-matchedMsgs))

			assert.DeepEqual(t, msgs, expectedSlots)
		})
	}

	cases := []struct {
		msgsCount     int
		seqset        string
		matchedMsgs   int
		expectedSlots []uint32
	}{
		{5, "1:*", 5, []uint32{}},
		{5, "*", 1, []uint32{1, 2, 3, 4}},
		{5, "1", 1, []uint32{2, 3, 4, 5}},
		{5, "2,1,5", 3, []uint32{3, 4}},
	}

	if os.Getenv("SHUFFLE_CASES") == "1" {
		rand.Shuffle(len(cases), func(i, j int) {
			cases[i], cases[j] = cases[j], cases[i]
		})
	}

	for _, case_ := range cases {
		testSlots(case_.msgsCount, case_.seqset, case_.matchedMsgs, case_.expectedSlots)
	}

	// Make sure backend returns seqnums, not UIDs.
	t.Run("Not UIDs", func(t *testing.T) {
		skipIfExcluded(t)
		
		conn := collectorConn{}

		mbox := getMbox(t, u, &conn)
		defer mbox.Close()
		createMsgs(t, mbox, u,6)

		conn.upds = nil

		seq, _ := imap.ParseSeqSet("1")
		assert.NilError(t, mbox.UpdateMessagesFlags(false, seq, imap.AddFlags, true, []string{imap.DeletedFlag}))
		assert.NilError(t, mbox.Expunge())

		conn.upds = nil
		
		msgs := makeMsgSlots(5)
		seq, _ = imap.ParseSeqSet("2,1,5")
		assert.NilError(t, mbox.UpdateMessagesFlags(false, seq, imap.AddFlags, true, []string{imap.DeletedFlag}))

		assert.NilError(t, mbox.Expunge())
		checkExpungeEvents(t, conn.upds, &msgs, uint32(2))

		assert.DeepEqual(t, msgs, []uint32{3, 4})
	})
}
