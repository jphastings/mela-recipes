gen:
	@go generate ./...
test:
	@go install github.com/tvastar/test/cmd/testmd@latest
	@testmd -o mela/README_test.go -pkg mela_test mela/README.md
ifeq ($(CI),"TRUE")
	@go test -json ./... > test-results.json
else
	@go test ./... && rm mela/README_test.go
endif
