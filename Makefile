compile_proto:
	protoc --plugin=$$(go env GOPATH)/bin/protoc-gen-go proto/*.proto --go_out=. --go_opt=module=github.com/brunokim/prol-go

$PHONY: compile_proto
