test:
	@testmd -o README_test.go -pkg mela_test README.md
	@go test ./... && rm README_test.go
