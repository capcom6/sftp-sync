project_name = sftp-sync

init:
	go mod download

run:
	go run .

test:
	go test -cover ./...

install:
	go install .

.PHONY: init run test install