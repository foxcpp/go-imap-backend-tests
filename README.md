go-imap-backend-tests
-----------------------

Blackbox tests for [go-imap] backends. 

The main intention of suite is to test basic RFC 3501 conformance.
Tests are developed in parallel with [go-imap-sql] so they reflect its
conformance.

### Tests

* IMAPUserDB interface tests
* Tests for mailbox management commands
* Tests for SEARCH and FETCH commands (for UID versions too) (ListMessages, SearchMessages)
* Tests for COPY/UID COPY commands (CopyMessages)
* Tests for STATUS command (Status)
* Tests for EXPUNGE command (Expunge)
* Tests for UPDATE command (SetMessagesFlags)
* Tests for unilateral updates (optional, backend.Updater interface)
* Test for UID monotonic increase
* Test for UIDVALIDITY/UIDNEXT change on mailbox rename
* APPENDLIMIT extension tests (optional, see [appendlimit.go][appendlimit.go] for interfaces)
* CHILDREN extension tests (optional, see [children/server.go][children/server.go] for interfaces)
* MOVE extension tests (optional) (MoveMessages)

### Incomplete RFC 3501 conformance

As this suite reflects state of go-imap-sql implementation, it may not test for
all requirements of IMAP specification.
There are known ignored cases:

* /NoSelect attribute and removal of mailboxes with children

### How to use

Tested backend must implement IMAPUsersDB interface.

Just call `testsuite.RunTests(t, newBackend, closeBackend)` from your backend (or
`backend_test`) package.  Each invocation of newBackend callback should provide
clean instance of backend (e.g. with empty storage, etc).  closeBackend will be
called for backend after usage. New instance is created for each test.

[go-imap]: https://github.com/emersion/go-imap
[go-imap-sql]: https://github.com/foxcpp/go-imap-sql
