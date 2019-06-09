package backendtests

import (
	"io"
	"io/ioutil"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-imap"
	"gotest.tools/assert"
)

// Our little torture test.
var encodedTestMsg = `From: "fox.cpp" <foxcpp@foxcpp.dev>
To: "fox.cpp" <foxcpp@foxcpp.dev>
Subject: =?utf-8?B?0J/RgNC+0LLQtdGA0LrQsCE=?=
Date: Sun, 09 Jun 2019 00:06:43 +0300
MIME-Version: 1.0
Message-ID: <a2aeb99e-52dd-40d3-b99f-1fdaad77ed98@foxcpp.dev>
Content-Type: text/plain; charset=utf-8; format=flowed
Content-Transfer-Encoding: quoted-printable

=D0=AD=D1=82=D0=BE=D1=82 =D1=82=D0=B5=D0=BA=D1=81=D1=82 =D0=B4=D0=BE=D0=BB=
=D0=B6=D0=B5=D0=BD =D0=B1=D1=8B=D1=82=D1=8C =D0=B7=D0=B0=D0=BA=D0=BE=D0=B4=
=D0=B8=D1=80=D0=BE=D0=B2=D0=B0=D0=BD =D0=B2 base64 =D0=B8=D0=BB=D0=B8 quote=
d-encoding.`

func Mailbox_FetchEncoded(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	u := getUser(t, b)
	defer u.Logout()
	mbox := getMbox(t, u)
	assert.NilError(t, mbox.CreateMessage([]string{}, time.Now(), strings.NewReader(encodedTestMsg)))
	seq, _ := imap.ParseSeqSet("1")

	t.Run("envelope", func(t *testing.T) {
		// https://tools.ietf.org/html/rfc3501#section-2.3.5
		// >A parsed representation of the [RFC-2822] header of the message.
		// It refers to RFC-2822 header, not MIME, meaning that it fields should
		// be decoded.

		ch := make(chan *imap.Message, 1)
		assert.NilError(t, mbox.ListMessages(false, seq, []imap.FetchItem{imap.FetchEnvelope}, ch))
		msg := <-ch

		assert.Equal(t, msg.Envelope.From[0].PersonalName, "fox.cpp", "PersonalName of From address is different (???)")
		assert.Equal(t, msg.Envelope.Subject, "=?utf-8?B?0J/RgNC+0LLQtdGA0LrQsCE=?=", "Subject field value is different (decoded?)")
	})
	t.Run("header subset", func(t *testing.T) {
		ch := make(chan *imap.Message, 1)
		assert.NilError(t, mbox.ListMessages(false, seq, []imap.FetchItem{imap.FetchItem("BODY.PEEK[HEADER.FIELDS (From Subject)]")}, ch))
		msg := <-ch

		var body io.Reader
		for _, lit := range msg.Body {
			body = lit
		}
		bodyBlob, err := ioutil.ReadAll(body)
		assert.NilError(t, err, "Literal ReadAll failed")

		t.Log(string(bodyBlob))

		// It might turn LF into CRLF, but whatever. We are focusing on checks
		// of header encoding here.
		assert.Check(t, strings.Contains(string(bodyBlob), `From: "fox.cpp" <foxcpp@foxcpp.dev>`), "Missing or different From field")
		assert.Check(t, strings.Contains(string(bodyBlob), `Subject: =?utf-8?B?0J/RgNC+0LLQtdGA0LrQsCE=?=`), "Missing or different Subject field")
	})
	t.Run("body subset", func(t *testing.T) {
		ch := make(chan *imap.Message, 1)
		assert.NilError(t, mbox.ListMessages(false, seq, []imap.FetchItem{imap.FetchItem("BODY.PEEK[]<360.2>")}, ch))
		msg := <-ch

		var body io.Reader
		for _, lit := range msg.Body {
			body = lit
		}
		bodyBlob, err := ioutil.ReadAll(body)
		assert.NilError(t, err, "Literal ReadAll failed")

		t.Log(string(bodyBlob))

		assert.Equal(t, string(bodyBlob), "E=", "Backend returns decoded or invalid BODY")
	})
}

func Mailbox_MatchEncoded(t *testing.T, newBack NewBackFunc, closeBack CloseBackFunc) {
	b := newBack()
	defer closeBack(b)
	u := getUser(t, b)
	defer u.Logout()
	mbox := getMbox(t, u)
	assert.NilError(t, mbox.CreateMessage([]string{}, time.Now(), strings.NewReader(encodedTestMsg)))

	t.Run("header", func(t *testing.T) {
		crit := imap.SearchCriteria{
			Header: textproto.MIMEHeader{"Subject": []string{"Проверка!"}},
		}
		seqs, err := mbox.SearchMessages(false, &crit)
		assert.NilError(t, err, "SearchMessages")
		assert.Equal(t, len(seqs), 1, "Not matched against decoded value")
	})
	t.Run("body", func(t *testing.T) {
		crit := imap.SearchCriteria{
			Text: []string{"или"},
		}
		seqs, err := mbox.SearchMessages(false, &crit)
		assert.NilError(t, err, "SearchMessages")
		assert.Equal(t, len(seqs), 1, "Not matched against decoded value")
	})
}
