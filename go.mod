module github.com/foxcpp/go-imap-backend-tests

go 1.12

require (
	github.com/emersion/go-imap v1.0.0-beta.4.0.20190504114255-4d5af3d05147
	github.com/emersion/go-imap-appendlimit v0.0.0-20190308131241-25671c986a6a
	github.com/emersion/go-imap-move v0.0.0-20180601155324-5eb20cb834bf
	github.com/emersion/go-message v0.11.1
	github.com/google/go-cmp v0.2.0
	github.com/pkg/errors v0.8.1 // indirect
	gotest.tools v2.2.0+incompatible
)

replace github.com/emersion/go-imap => github.com/foxcpp/go-imap v1.0.0-beta.1.0.20201001193006-5a1d05e53e2c
