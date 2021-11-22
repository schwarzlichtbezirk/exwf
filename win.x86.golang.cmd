@echo off
go env -w GOOS=windows GOARCH=386
cd /d %GOPATH%\src\github.com\schwarzlichtbezirk\exwf
go build -o %GOPATH%/bin/exwf.x86.exe -v github.com/schwarzlichtbezirk/exwf/cmd
