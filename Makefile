.PHONY: pb

pb: ## Build protobuf
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install github.com/favadi/protoc-go-inject-tag@latest
	@find . -type f -name '*.pb.go' -exec rm -rf {} \;
	@find . -type f -name '*.proto' -exec protoc --proto_path=. --go_out=paths=source_relative:. {} \;
	@find . -type f -name '*.pb.go' -exec protoc-go-inject-tag -input={} \;