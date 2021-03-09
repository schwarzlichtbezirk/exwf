@echo off
go env -w GOOS=windows GOARCH=amd64
go build -o %GOPATH%\bin\exwf.x64.exe -v github.com/schwarzlichtbezirk/exwf/cmd
