@ECHO OFF

SET APP=NFA
SET TARGET=CustomProtoServer.exe

SET GOOS=windows
SET GOARCH=amd64

echo Compile JCE files
%GOPATH%\bin\tars2go --outdir %cd%\vendor CustomProto.jce

echo Building
go build -o "%TARGET%"
"%TARGET%" --config=CustomProtoServer.conf