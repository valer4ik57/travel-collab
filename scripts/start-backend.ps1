Set-Location $PSScriptRoot\..\backend
go mod tidy
go run ./cmd/main.go
