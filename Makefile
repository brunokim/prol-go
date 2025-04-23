compile_proto:
	protoc --plugin=$$(go env GOPATH)/bin/protoc-gen-go proto/*.proto --go_out=. --go_opt=paths=source_relative

$PHONY: compile_proto
