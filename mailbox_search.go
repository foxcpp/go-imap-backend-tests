package backendtests

import (
	"net/textproto"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-imap"
	"gotest.tools/assert"
)

// Based on tests for go-imap/backend/backendutil.
// https://github.com/emersion/go-imap/blob/v1/backend/backendutil
//
// Intended for backends using custom search implementation.

var testInternalDate = time.Unix(1483997966, 0)

var matchTests = []struct {
	flags    []string
	date     time.Time
	criteria *imap.SearchCriteria
	res      bool
}{
	{
		criteria: &imap.SearchCriteria{
			Header: textproto.MIMEHeader{"From": {"Mitsuha"}},
		},
		res: true,
	},
	{
		criteria: &imap.SearchCriteria{
			Header: textproto.MIMEHeader{"To": {"Mitsuha"}},
		},
		res: false,
	},
	{
		criteria: &imap.SearchCriteria{SentBefore: testDate.Add(48 * time.Hour)},
		res:      true,
	},
	{
		criteria: &imap.SearchCriteria{
			Not: []*imap.SearchCriteria{{SentSince: testDate.Add(48 * time.Hour)}},
		},
		res: true,
	},
	{
		criteria: &imap.SearchCriteria{
			Not: []*imap.SearchCriteria{{Body: []string{"name"}}},
		},
		res: false,
	},
	{
		criteria: &imap.SearchCriteria{
			Text: []string{"name"},
		},
		res: true,
	},
	{
		criteria: &imap.SearchCriteria{
			Or: [][2]*imap.SearchCriteria{{
				{Text: []string{"i'm not in the text"}},
				{Body: []string{"i'm not in the body"}},
			}},
		},
		res: false,
	},
	{
		criteria: &imap.SearchCriteria{
			Header: textproto.MIMEHeader{"Message-Id": {"42@example.org"}},
		},
		res: true,
	},
	{
		criteria: &imap.SearchCriteria{
			Header: textproto.MIMEHeader{"Message-Id": {"43@example.org"}},
		},
		res: false,
	},
	{
		criteria: &imap.SearchCriteria{
			Header: textproto.MIMEHeader{"Message-Id": {""}},
		},
		res: true,
	},
	{
		criteria: &imap.SearchCriteria{
			Header: textproto.MIMEHeader{"Reply-To": {""}},
		},
		res: false,
	},
	{
		criteria: &imap.SearchCriteria{
			Larger: 10,
		},
		res: true,
	},
	{
		criteria: &imap.SearchCriteria{
			Smaller: 10,
		},
		res: false,
	},
	{
		criteria: &imap.SearchCriteria{
			Header: textproto.MIMEHeader{"Subject": {"your"}},
		},
		res: true,
	},
	{
		criteria: &imap.SearchCriteria{
			Header: textproto.MIMEHeader{"Subject": {"Taki"}},
		},
		res: false,
	},
	{
		flags: []string{imap.SeenFlag},
		criteria: &imap.SearchCriteria{
			WithFlags:    []string{imap.SeenFlag},
			WithoutFlags: []string{imap.FlaggedFlag},
		},
		res: true,
	},
	{
		flags: []string{imap.SeenFlag},
		criteria: &imap.SearchCriteria{
			WithFlags:    []string{imap.DraftFlag},
			WithoutFlags: []string{imap.FlaggedFlag},
		},
		res: false,
	},
	{
		flags: []string{imap.SeenFlag, imap.FlaggedFlag},
		criteria: &imap.SearchCriteria{
			WithFlags:    []string{imap.SeenFlag},
			WithoutFlags: []string{imap.FlaggedFlag},
		},
		res: false,
	},
	{
		flags: []string{imap.SeenFlag, imap.FlaggedFlag},
		criteria: &imap.SearchCriteria{
			Or: [][2]*imap.SearchCriteria{{
				{WithFlags: []string{imap.DraftFlag}},
				{WithoutFlags: []string{imap.SeenFlag}},
			}},
		},
		res: false,
	},
	{
		flags: []string{imap.SeenFlag, imap.FlaggedFlag},
		criteria: &imap.SearchCriteria{
			Not: []*imap.SearchCriteria{
				{WithFlags: []string{imap.SeenFlag}},
			},
		},
		res: false,
	},
	{
		criteria: &imap.SearchCriteria{
			Or: [][2]*imap.SearchCriteria{{
				{
					Uid: new(imap.SeqSet),
					Not: []*imap.SearchCriteria{{SeqNum: new(imap.SeqSet)}},
				},
				{
					SeqNum: new(imap.SeqSet),
				},
			}},
		},
		res: false,
	},
	{
		criteria: &imap.SearchCriteria{
			Or: [][2]*imap.SearchCriteria{{
				{
					Uid: &imap.SeqSet{Set: []imap.Seq{{2, 2}}},
					Not: []*imap.SearchCriteria{{SeqNum: new(imap.SeqSet)}},
				},
				{
					SeqNum: new(imap.SeqSet),
				},
			}},
		},
		res: true,
	},
	{
		criteria: &imap.SearchCriteria{
			Or: [][2]*imap.SearchCriteria{{
				{
					Uid: &imap.SeqSet{Set: []imap.Seq{{2, 2}}},
					Not: []*imap.SearchCriteria{{
						SeqNum: &imap.SeqSet{Set: []imap.Seq{imap.Seq{1, 1}}},
					}},
				},
				{
					SeqNum: new(imap.SeqSet),
				},
			}},
		},
		res: false,
	},
	{
		criteria: &imap.SearchCriteria{
			Or: [][2]*imap.SearchCriteria{{
				{
					Uid: &imap.SeqSet{Set: []imap.Seq{{2, 2}}},
					Not: []*imap.SearchCriteria{{
						SeqNum: &imap.SeqSet{Set: []imap.Seq{{1, 1}}},
					}},
				},
				{
					SeqNum: &imap.SeqSet{Set: []imap.Seq{{1, 1}}},
				},
			}},
		},
		res: true,
	},
	{
		date: testInternalDate,
		criteria: &imap.SearchCriteria{
			Or: [][2]*imap.SearchCriteria{{
				{
					Since: testInternalDate.Add(48 * time.Hour),
					Not: []*imap.SearchCriteria{{
						Since: testInternalDate.Add(48 * time.Hour),
					}},
				},
				{
					Before: testInternalDate.Add(-48 * time.Hour),
				},
			}},
		},
		res: false,
	},
	{
		date: testInternalDate,
		criteria: &imap.SearchCriteria{
			Or: [][2]*imap.SearchCriteria{{
				{
					Since: testInternalDate.Add(-48 * time.Hour),
					Not: []*imap.SearchCriteria{{
						Since: testInternalDate.Add(48 * time.Hour),
					}},
				},
				{
					Before: testInternalDate.Add(-48 * time.Hour),
				},
			}},
		},
		res: true,
	},
	{
		date: testInternalDate,
		criteria: &imap.SearchCriteria{
			Or: [][2]*imap.SearchCriteria{{
				{
					Since: testInternalDate.Add(-48 * time.Hour),
					Not: []*imap.SearchCriteria{{
						Since: testInternalDate.Add(-48 * time.Hour),
					}},
				},
				{
					Before: testInternalDate.Add(-48 * time.Hour),
				},
			}},
		},
		res: false,
	},
	{
		date: testInternalDate,
		criteria: &imap.SearchCriteria{
			Or: [][2]*imap.SearchCriteria{{
				{
					Since: testInternalDate.Add(-48 * time.Hour),
					Not: []*imap.SearchCriteria{{
						Since: testInternalDate.Add(-48 * time.Hour),
					}},
				},
				{
					Before: testInternalDate.Add(48 * time.Hour),
				},
			}},
		},
		res: true,
	},
}

func Mailbox_SearchMessages(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	u := getUser(t, b)
	defer assert.NilError(t, u.Logout())

	for i, test := range matchTests {
		test := test
		t.Run("Crit "+strconv.Itoa(i+1), func(t *testing.T) {
			skipIfExcluded(t)
			mbox := getMbox(t, u)

			// Create a message and delete it to make sure test message will have seqnum=1 and uid=2.
			assert.NilError(t, mbox.CreateMessage(test.flags, test.date, strings.NewReader(testMailString)))
			assert.NilError(t, mbox.UpdateMessagesFlags(false, &imap.SeqSet{Set: []imap.Seq{{1, 1}}}, imap.AddFlags, []string{imap.DeletedFlag}))
			assert.NilError(t, mbox.Expunge())

			assert.NilError(t, mbox.CreateMessage(test.flags, test.date, strings.NewReader(testMailString)))

			t.Run("seq", func(t *testing.T) {
				res, err := mbox.SearchMessages(false, test.criteria)
				assert.NilError(t, err)
				if test.res {
					if !assert.Check(t, len(res) == 1 && res[0] == 1, "Criteria not matched when expected") {
						t.Logf("Criteria: %+v\n", test.criteria)
						t.Logf("Res: %+v\n", res)
					}
				} else {
					if !assert.Check(t, len(res) == 0, "Criteria matched when not expected") {
						t.Logf("Criteria: %+v\n", test.criteria)
						t.Logf("Res: %+v\n", res)
					}
				}
			})
			t.Run("uid", func(t *testing.T) {
				res, err := mbox.SearchMessages(true, test.criteria)
				assert.NilError(t, err)
				if test.res {
					if !assert.Check(t, len(res) == 1 && res[0] == 2, "Criteria not matched when expected") {
						t.Logf("Criteria: %+v\n", test.criteria)
						t.Logf("Res: %+v\n", res)
					}
				} else {
					if !assert.Check(t, len(res) == 0, "Criteria matched when not expected") {
						t.Logf("Criteria: %+v\n", test.criteria)
						t.Logf("Res: %+v\n", res)
					}
				}
			})
		})
	}
}
