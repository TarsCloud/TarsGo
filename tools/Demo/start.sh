#!/bin/bash
tars2go -outdir=. *.tars
go build
./_SERVER_ --config=_SERVER_.conf
