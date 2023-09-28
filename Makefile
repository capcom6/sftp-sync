project_name = sftp-sync

init:
	go mod download

run:
	go run cmd/$(project_name)/main.go

test:
	go test -cover ./...

install:
	go install cmd/$(project_name)/main.go

.PHONY: init run test install