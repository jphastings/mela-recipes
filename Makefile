test:
	@echo "Testing library:"
	@go test
	@echo "Testing code in README.md:"
	@go install github.com/tvastar/test/cmd/testmd@latest
	@testmd -o README_test.go -pkg mela_test README.md && go test README_test.go; rm README_test.go
