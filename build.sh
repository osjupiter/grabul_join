#!/bin/bash
go get github.com/ChimeraCoder/anaconda
go get github.com/lxn/walk
go get github.com/akavel/rsrc
go get github.com/mattn/go-ieproxy/global
rsrc -manifest test.manifest -o rsrc.syso

go build
