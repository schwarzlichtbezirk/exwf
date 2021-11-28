@echo off
go env -w GOOS=windows GOARCH=amd64
cd /d %GOPATH%\src\github.com\schwarzlichtbezirk\exwf
go build -o %GOPATH%/bin/exwf.x64.exe -v ./cmd
