test:
	@go install github.com/tvastar/test/cmd/testmd@latest
	@testmd -o README_test.go -pkg mela_test README.md
ifeq ($(CI),"TRUE")
	@go test -json ./... > test-results.json
else
	@go test ./... && rm README_test.go
endif
