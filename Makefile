build-cmd:
	go build -ldflags "-X 'cli-aio/cmd/version.Version=$$(git describe --tags --always --dirty)' -X 'cli-aio/cmd/version.BuildTime=$$(date -u +%Y-%m-%dT%H:%M:%SZ)' -X 'cli-aio/cmd/version.GitCommit=$$(git rev-parse --short HEAD)'" -o aio .
	mv ./aio /usr/local/bin/