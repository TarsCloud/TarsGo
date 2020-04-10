@ECHO OFF

SET APP=NFA
SET TARGET=CustomProtoServer

SET TARS2GO="%GOPATH%\bin\tars2go"
SET TARGZIP="%GOPATH%\bin\targzip"
SET PATCH_TOOL=dpatch.exe

rem Remember old GOOS
SET Last_GOOS=%GOOS%

SET GOOS=linux
SET GOARCH=amd64

go build -o "%TARGET%"

rem Set GOOS back to the previous one
SET GOOS=%Last_GOOS%

md "tartmpdir/%TARGET%"
copy "%TARGET%" "tartmpdir/%TARGET%"
cd tartmpdir
%TARGZIP%  "%TARGET%"  ..\\"%TARGET%.tgz"
cd ..
rd /S /Q tartmpdir

%PATCH_TOOL% --env test --moduleType taf --application "%APP%" --server "%TARGET%" --tgz "%TARGET%".tgz --user _DEVELOPER_