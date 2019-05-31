package backendtests

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-imap"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

// Mostly taken from go-imap/backend/backendutil tests.
// https://github.com/emersion/go-imap/blob/v1/backend/backendutil
//
// This tests block is intended to test backend's ability to return
// correct body sections if it implements some advanced indexing
// for them to increase performance.

const testBodyString = "--message-boundary\r\n" +
	testAltHeaderString +
	"\r\n--b2\r\n" +
	testTextString +
	"\r\n--b2\r\n" +
	testHTMLString +
	"\r\n--b2--\r\n" +
	"\r\n--message-boundary\r\n" +
	testAttachmentString +
	"\r\n--message-boundary--\r\n"

const testMailString = testHeaderString + testBodyString

var bodyTests = []struct {
	section string
	body    string
}{
	{
		section: "BODY[]",
		body:    testMailString,
	},
	{
		section: "BODY[1.1]",
		body:    testTextBodyString,
	},
	{
		section: "BODY[1.2]",
		body:    testHTMLBodyString,
	},
	{
		section: "BODY[2]",
		body:    testAttachmentBodyString,
	},
	{
		section: "BODY[HEADER]",
		body:    testHeaderString,
	},
	{
		section: "BODY[HEADER.FIELDS (From To)]",
		body:    testHeaderFromToString,
	},
	{
		section: "BODY[HEADER.FIELDS.NOT (From To)]",
		body:    testHeaderNoFromToString,
	},
	{
		section: "BODY[HEADER.FIELDS (From To)]<0.2>",
		body:    testHeaderFromToString[:2],
	},
	{
		section: "BODY[HEADER.FIELDS.NOT (From To)]<0.2>",
		body:    testHeaderNoFromToString[:2],
	},
	{
		section: "BODY[1.1.HEADER]",
		body:    testTextHeaderString,
	},
	{
		section: "BODY[1.1.HEADER.FIELDS (Content-Type)]",
		body:    testTextContentTypeString,
	},
	{
		section: "BODY[1.1.HEADER.FIELDS.NOT (Content-Type)]",
		body:    testTextNoContentTypeString,
	},
	{
		section: "BODY[1.1.HEADER.FIELDS (Content-Type)]<0.2>",
		body:    testTextContentTypeString[:2],
	},
	{
		section: "BODY[1.1.HEADER.FIELDS.NOT (Content-Type)]<0.2>",
		body:    testTextNoContentTypeString[:2],
	},
	{
		section: "BODY[2.HEADER]",
		body:    testAttachmentHeaderString,
	},
	{
		section: "BODY[2.MIME]",
		body:    testAttachmentHeaderString,
	},
	{
		section: "BODY[TEXT]",
		body:    testBodyString,
	},
	{
		section: "BODY[1.1.TEXT]",
		body:    testTextBodyString,
	},
	{
		section: "BODY[2.TEXT]",
		body:    testAttachmentBodyString,
	},
	{
		section: "BODY[2.1]",
		body:    "",
	},
	{
		section: "BODY[3]",
		body:    "",
	},
	{
		section: "BODY[2.TEXT]<0.9>",
		body:    testAttachmentBodyString[:9],
	},
}

func Mailbox_ListMessages_BodyPeek(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	u := getUser(t, b)
	defer assert.NilError(t, u.Logout())

	t.Run("without PEEK", func(t *testing.T) {
		mbox := getMbox(t, u)

		date := time.Now()
		err := mbox.CreateMessage([]string{"$Test1", "$Test2"}, date, strings.NewReader(testMailString))
		assert.NilError(t, err)

		seq, _ := imap.ParseSeqSet("1")
		ch := make(chan *imap.Message, 10)
		assert.NilError(t, mbox.ListMessages(false, seq, []imap.FetchItem{"BODY[]"}, ch))
		msg := <-ch

		// Changed flag should be included in fetch.
		if _, ok := msg.Items[imap.FetchFlags]; !ok {
			t.Fatal("flags are not returned when changed by BODY[]")
		}
		containsSeen := false
		for _, flag := range msg.Flags {
			if flag == imap.SeenFlag {
				containsSeen = true
			}
		}
		if !containsSeen {
			t.Fatal("\\Seen flag is not set/returned when BODY[] is fetched")
		}
	})
	t.Run("with PEEK", func(t *testing.T) {
		mbox := getMbox(t, u)

		date := time.Now()
		err := mbox.CreateMessage([]string{"$Test1", "$Test2"}, date, strings.NewReader(testMailString))
		assert.NilError(t, err)

		seq, _ := imap.ParseSeqSet("1")
		ch := make(chan *imap.Message, 10)
		assert.NilError(t, mbox.ListMessages(false, seq, []imap.FetchItem{"BODY.PEEK[]", imap.FetchFlags}, ch))
		msg := <-ch

		containsSeen := false
		for _, flag := range msg.Flags {
			if flag == imap.SeenFlag {
				containsSeen = true
			}
		}
		if containsSeen {
			t.Fatal("\\Seen flag is set when BODY.PEEK[] is fetched")
		}
	})
	t.Run("non-body", func(t *testing.T) {
		mbox := getMbox(t, u)

		date := time.Now()
		err := mbox.CreateMessage([]string{"$Test1", "$Test2"}, date, strings.NewReader(testMailString))
		assert.NilError(t, err)

		seq, _ := imap.ParseSeqSet("1")
		ch := make(chan *imap.Message, 10)
		assert.NilError(t, mbox.ListMessages(false, seq, []imap.FetchItem{imap.FetchUid, imap.FetchFlags}, ch))
		msg := <-ch

		containsSeen := false
		for _, flag := range msg.Flags {
			if flag == imap.SeenFlag {
				containsSeen = true
			}
		}
		if containsSeen {
			t.Fatal("\\Seen flag is set when non-body item is fetched")
		}
	})
}

func Mailbox_ListMessages_Body(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	u := getUser(t, b)
	defer assert.NilError(t, u.Logout())
	mbox := getMbox(t, u)

	date := time.Now()
	err := mbox.CreateMessage([]string{"$Test1", "$Test2"}, date, strings.NewReader(testMailString))
	assert.NilError(t, err)

	seq, _ := imap.ParseSeqSet("1")

	for _, test := range bodyTests {
		test := test
		t.Run(test.section, func(t *testing.T) {
			skipIfExcluded(t)

			ch := make(chan *imap.Message, 10)
			err := mbox.ListMessages(false, seq, []imap.FetchItem{imap.FetchItem(test.section)}, ch)
			if test.body == "" && err != nil {
				return
			}
			assert.NilError(t, err, "mbox.ListMessages")

			assert.Assert(t, is.Len(ch, 1), "Wrong number of messages returned")
			msg := <-ch
			assert.Equal(t, msg.SeqNum, uint32(1))

			for k, literal := range msg.Body {
				if k.FetchItem() != imap.FetchItem(test.section) {
					t.Fatal("Unexpected body section returned:", k.FetchItem())
				}

				body, err := ioutil.ReadAll(literal)
				assert.NilError(t, err, "Failed to read body section")

				assert.DeepEqual(t, test.body, string(body))
			}

		})
	}
}

var testBodyStructure = &imap.BodyStructure{
	MIMEType:    "multipart",
	MIMESubType: "mixed",
	Params:      map[string]string{"boundary": "message-boundary"},
	Parts: []*imap.BodyStructure{
		{
			MIMEType:    "multipart",
			MIMESubType: "alternative",
			Params:      map[string]string{"boundary": "b2"},
			Extended:    true,
			Parts: []*imap.BodyStructure{
				{
					MIMEType:          "text",
					MIMESubType:       "plain",
					Params:            map[string]string{},
					Extended:          true,
					Disposition:       "inline",
					DispositionParams: map[string]string{},
				},
				{
					MIMEType:          "text",
					MIMESubType:       "html",
					Params:            map[string]string{},
					Extended:          true,
					Disposition:       "inline",
					DispositionParams: map[string]string{},
				},
			},
		},
		{
			MIMEType:          "text",
			MIMESubType:       "plain",
			Params:            map[string]string{},
			Extended:          true,
			Disposition:       "attachment",
			DispositionParams: map[string]string{"filename": "note.txt"},
		},
	},
	Extended: true,
}

var testEnvelope = &imap.Envelope{
	Date: testDate,
	From: []*imap.Address{
		&imap.Address{
			PersonalName: "Mitsuha Miyamizu",
			MailboxName:  "mitsuha.miyamizu",
			HostName:     "example.org",
		},
	},
	To: []*imap.Address{
		&imap.Address{
			PersonalName: "Taki Tachibana",
			MailboxName:  "taki.tachibana",
			HostName:     "example.org",
		},
	},
	Subject:   "Your Name.",
	MessageId: "42@example.org",
	Sender:    []*imap.Address{},
	Cc:        []*imap.Address{},
	Bcc:       []*imap.Address{},
	ReplyTo:   []*imap.Address{},
}

func stripEnvelope(env *imap.Envelope) {
	if env.From == nil {
		env.From = []*imap.Address{}
	}
	if env.To == nil {
		env.To = []*imap.Address{}
	}
	if env.Sender == nil {
		env.Sender = []*imap.Address{}
	}
	if env.Cc == nil {
		env.Cc = []*imap.Address{}
	}
	if env.Bcc == nil {
		env.Bcc = []*imap.Address{}
	}
	if env.ReplyTo == nil {
		env.ReplyTo = []*imap.Address{}
	}
}

func stripBodyStructure(bs *imap.BodyStructure) {
	if bs.Params == nil {
		bs.Params = map[string]string{}
	}
	if bs.Parts == nil {
		bs.Parts = []*imap.BodyStructure{}
	}
	if bs.DispositionParams == nil {
		bs.DispositionParams = map[string]string{}
	}
	if bs.Language == nil {
		bs.Language = []string{}
	}
	if bs.Location == nil {
		bs.Location = []string{}
	}

	for _, part := range bs.Parts {
		stripBodyStructure(part)
	}
}

func Mailbox_ListMessages_Meta(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	u := getUser(t, b)
	defer assert.NilError(t, u.Logout())
	mbox := getMbox(t, u)
	createMsgs(t, mbox, 1)

	t.Run("fetch bodystruct", func(t *testing.T) {
		skipIfExcluded(t)

		seq, _ := imap.ParseSeqSet("1")
		ch := make(chan *imap.Message, 10)
		assert.NilError(t, mbox.ListMessages(false, seq, []imap.FetchItem{imap.FetchBodyStructure}, ch))
		assert.Assert(t, is.Len(ch, 1), "Wrong number of messages returned")
		msg := <-ch
		assert.Equal(t, msg.SeqNum, uint32(1))

		stripBodyStructure(msg.BodyStructure)
		stripBodyStructure(testBodyStructure)
		assert.DeepEqual(t, msg.BodyStructure, testBodyStructure)
	})

	t.Run("fetch envelope", func(t *testing.T) {
		skipIfExcluded(t)

		seq, _ := imap.ParseSeqSet("1")
		ch := make(chan *imap.Message, 10)
		assert.NilError(t, mbox.ListMessages(false, seq, []imap.FetchItem{imap.FetchEnvelope}, ch))
		assert.Assert(t, is.Len(ch, 1), "Wrong number of messages returned")
		msg := <-ch

		stripEnvelope(msg.Envelope)
		stripEnvelope(testEnvelope)
		assert.DeepEqual(t, msg.Envelope, testEnvelope)
	})
}

func Mailbox_ListMessages_Multi(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	u := getUser(t, b)
	defer assert.NilError(t, u.Logout())
	mbox := getMbox(t, u)
	createMsgs(t, mbox, 1)

	t.Run("fetch uid,body[]", func(t *testing.T) {
		skipIfExcluded(t)

		seq, _ := imap.ParseSeqSet("1")
		ch := make(chan *imap.Message, 10)
		assert.NilError(t, mbox.ListMessages(false, seq, []imap.FetchItem{imap.FetchUid, imap.FetchItem("BODY[]")}, ch))
		assert.Assert(t, is.Len(ch, 1), "Wrong number of messages returned")
		msg := <-ch

		assert.Check(t, is.Equal(msg.Uid, uint32(1)), "UID mismatch")
		assert.Check(t, is.Len(msg.Body, 1), "Wrong amount of body sections")
		for _, v := range msg.Body {
			blob, err := ioutil.ReadAll(v)
			assert.NilError(t, err, "Body ReadAll failed")
			assert.Check(t, is.DeepEqual(testMailString, string(blob)))
		}
	})
	t.Run("fetch uid,body[header]", func(t *testing.T) {
		skipIfExcluded(t)

		seq, _ := imap.ParseSeqSet("1")
		ch := make(chan *imap.Message, 10)
		assert.NilError(t, mbox.ListMessages(false, seq, []imap.FetchItem{imap.FetchUid, imap.FetchItem("BODY[HEADER]")}, ch))
		assert.Assert(t, is.Len(ch, 1), "Wrong number of messages returned")
		msg := <-ch

		assert.Check(t, is.Equal(msg.Uid, uint32(1)), "UID mismatch")
		assert.Check(t, is.Len(msg.Body, 1), "Wrong amount of body sections")
		for _, v := range msg.Body {
			blob, err := ioutil.ReadAll(v)
			assert.NilError(t, err, "Body ReadAll failed")
			assert.Check(t, is.DeepEqual(testHeaderString, string(blob)))
		}
	})
}
