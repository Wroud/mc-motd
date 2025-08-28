module github.com/wroud/mc-motd

go 1.24.4

require (
	github.com/itzg/go-flagsfiller v1.16.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/text v0.28.0
)

// go-kit pulls in old, ambiguous package
// github.com/go-kit/kit@v0.13.0 google.golang.org/genproto@v0.0.0-20210917145530-b395a37504d4
// ambiguous with
// google.golang.org/genproto/googleapis/api
// google.golang.org/genproto/googleapis/rpc
exclude google.golang.org/genproto v0.0.0-20210917145530-b395a37504d4

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/uuid v1.6.0
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/sys v0.33.0 // indirect
)
