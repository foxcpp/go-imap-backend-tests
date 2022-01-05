module github.com/foxcpp/go-imap-backend-tests

go 1.12

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emersion/go-imap v1.0.0-beta.4.0.20190504114255-4d5af3d05147
	github.com/emersion/go-message v0.15.0
	github.com/google/go-cmp v0.2.0
	github.com/martinlindhe/base36 v1.0.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/stretchr/testify v1.3.0 // indirect
	gotest.tools v2.2.0+incompatible
)

replace github.com/emersion/go-imap => github.com/foxcpp/go-imap v1.0.0-beta.1.0.20220105164802-1e767d4cfd62
