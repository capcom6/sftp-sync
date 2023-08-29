project_name = sftp-sync

init:
	go mod download

run:
	go run cmd/$(project_name)/main.go

test:
	go test ./...

.PHONY: init run test